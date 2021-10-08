import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.MIXIN_BASE_URL": "https://mixin-api.zeromesh.net",
    "process.env.RED_PACKET_ID": "1ab1f241-b809-4790-bcfd-a1779bb1d313",
    "process.env.SERVER_URL": "http://192.168.0.106:7001",
    "process.env.LIVE_REPLAY_URL":
      "https://super-group-cdn.mixinbots.com/live-replay/",
  },
})
