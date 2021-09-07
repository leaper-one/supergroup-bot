import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "en",
    "process.env.MIXIN_BASE_URL": "https://api.mixin.one",
    "process.env.RED_PACKET_ID": "70b94e54-8f75-41f5-91e2-12522112ee71",
    "process.env.SERVER_URL": "mixin.group",
    "process.env.LIVE_REPLAY_URL": "https://supergroup-cdn.mixin.group/live-replay/",
  },
})
