module.exports = {
  apps: [
    {
      name: "frontend-web",
      script: "node_modules/next/dist/bin/next",
      args: `start -p ${process.env.PORT || 3000} -H ${process.env.HOSTNAME || "0.0.0.0"}`,
      env: {
        NODE_ENV: "production",
      },
    },
  ],
};
