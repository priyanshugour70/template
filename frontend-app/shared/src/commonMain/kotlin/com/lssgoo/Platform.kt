package com.lssgoo

interface Platform {
    val name: String
}

expect fun getPlatform(): Platform