import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "en",
    "process.env.MIXIN_BASE_URL": "https://mixin-api.zeromesh.net",
    "process.env.RED_PACKET_URL": "http://192.168.2.237:8080",
    "process.env.RED_PACKET_ID": "",
    "process.env.SERVER_URL": "mixinbots.com",
    "process.env.LIVE_REPLAY_URL": "https://super-group-cdn.mixinbots.com/live-replay/",
  },
})
