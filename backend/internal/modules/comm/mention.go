package comm

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// mentionPattern recognises Slack-style @-mentions:
//   - @here / @channel / @everyone   → broadcast keywords (lowercase only)
//   - @<identifier>                  → user lookup; identifier is what the user
//                                      sees in the composer's autocomplete and
//                                      will typically be a username, email
//                                      local-part, or first.last
//
// The pattern requires whitespace OR start-of-string before the @ so emails
// inside the body ("ping nikhil@example.com") never get treated as mentions.
var mentionPattern = regexp.MustCompile(`(?m)(^|\s)@([A-Za-z][A-Za-z0-9._-]{1,63})\b`)

// RawMention is a parser-only value, BEFORE the identifier is resolved to a
// concrete user. The service performs the lookup against the org's member
// list and turns these into MessageMention rows.
type RawMention struct {
	// Token is the matched identifier WITHOUT the @ prefix, lower-cased.
	Token string
	// Type categorises the token: 'broadcast' for here/channel/everyone, 'user'
	// otherwise. Broadcast tokens skip the user-lookup step entirely.
	Type string
	// Index is the byte offset of the '@' in the original body. Lets the
	// frontend highlight the exact range later without re-parsing.
	Index int
}

const (
	rawMentionUser      = "user"
	rawMentionBroadcast = "broadcast"
)

var broadcastTokens = map[string]string{
	"here":     "here",
	"channel":  "channel",
	"everyone": "everyone",
}

// ParseMentions walks a message body and returns every @-mention in order of
// appearance. Duplicates are de-duped (lower-cased token) so spamming
// "@priyanshu @priyanshu @priyanshu" only generates ONE notification.
//
// The function is deliberately permissive about identifiers: it doesn't know
// what counts as a "valid user". The service layer does the lookup and quietly
// drops tokens that don't resolve — surfacing a "user not found" error per
// mention would punish honest typos.
func ParseMentions(body string) []RawMention {
	matches := mentionPattern.FindAllStringSubmatchIndex(body, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]RawMention, 0, len(matches))
	for _, m := range matches {
		// m = [matchStart, matchEnd, leadStart, leadEnd, tokenStart, tokenEnd]
		tokStart, tokEnd := m[4], m[5]
		atIdx := tokStart - 1 // back up over '@'
		raw := strings.ToLower(body[tokStart:tokEnd])
		if _, dup := seen[raw]; dup {
			continue
		}
		seen[raw] = struct{}{}
		rm := RawMention{Token: raw, Index: atIdx}
		if _, ok := broadcastTokens[raw]; ok {
			rm.Type = rawMentionBroadcast
		} else {
			rm.Type = rawMentionUser
		}
		out = append(out, rm)
	}
	return out
}

// BuildMentionsForMessage takes the raw parser output + the resolved
// identifier→user map and constructs MessageMention rows ready to be
// inserted. The service does the lookup; this helper only stitches.
//
// Broadcast tokens (here/channel/everyone) produce a single mention each with
// target_user_id = NULL. The mention-trigger code in notify.go expands
// 'channel'/'everyone' into actual recipients at notification time, so we
// don't bloat the mentions table with N rows for an N-member channel.
func BuildMentionsForMessage(messageID uuid.UUID, raw []RawMention, resolved map[string]uuid.UUID) []MessageMention {
	if len(raw) == 0 {
		return nil
	}
	out := make([]MessageMention, 0, len(raw))
	for _, m := range raw {
		mention := MessageMention{
			MessageID:   messageID,
			IndexInBody: m.Index,
		}
		if m.Type == rawMentionBroadcast {
			mention.MentionType = m.Token // here | channel | everyone
			out = append(out, mention)
			continue
		}
		// User token: drop silently if it didn't resolve to anyone in the org.
		uid, ok := resolved[m.Token]
		if !ok {
			continue
		}
		mention.MentionType = "user"
		mention.TargetUserID = &uid
		out = append(out, mention)
	}
	return out
}

// candidateIdentifiers returns the strings a parser token might match against
// in the user/membership lookup. The service builds a map keyed by these
// strings (lower-cased) against the user_id, then passes it to
// BuildMentionsForMessage.
//
// Order matters only for documentation: the service produces a single map and
// any collision (e.g. two users with the same first name) goes to whoever was
// hydrated first — same as Slack's "first match wins" behaviour.
func candidateIdentifiers(displayName, firstName, lastName, email, username string) []string {
	out := make([]string, 0, 4)
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, " ", ".")
		if s == "" {
			return
		}
		out = append(out, s)
	}
	add(username)
	add(displayName)
	if firstName != "" && lastName != "" {
		add(firstName + "." + lastName)
	}
	add(firstName)
	// Email local-part — ignore domain so two `priyanshu@*` users from
	// different domains don't collide. This is best-effort; admins should
	// set usernames if they want predictable mentions.
	if at := strings.Index(email, "@"); at > 0 {
		add(email[:at])
	}
	return out
}
