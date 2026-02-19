module.exports = {
  apps: [{
    name: "websocket-server",
    script: "dist/main.js",
    instances: 1, // whatsapp-web.js NO es cluster-safe
    exec_mode: "fork",
    max_memory_restart: "1G",
    node_args: [
      "--max-old-space-size=384",
      "--trace-warnings"
    ],
    env: {
      NODE_ENV: "production"
    },
    restart_delay: 5000,
    exp_backoff_restart_delay: 1000
  }]
}
