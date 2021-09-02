import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "en",
    "process.env.MIXIN_BASE_URL": "https://mixin-api.zeromesh.net",
    "process.env.RED_PACKET_URL": "http://192.168.2.237:8080",
    "process.env.RED_PACKET_ID": "4b85b71a-2b06-4809-a5e0-399680483dcd",
    "process.env.SERVER_URL": "mixinbots.com",
    "process.env.LIVE_REPLAY_URL": "https://super-group-cdn.mixinbots.com/live-replay/",
  },
})
