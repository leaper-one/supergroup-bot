import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.RED_PACKET_URL": "https://red-api.mixinbots.com",
    "process.env.SERVER_URL": "mixinbots.com",
    "process.env.LIVE_REPLAY_URL": "https://super-group-cdn.mixinbots.com/live-replay/",
  },
})
