import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "en",
    "process.env.RED_PACKET_URL": "https://redpacket.mixin.zone",
    "process.env.SERVER_URL": "mixin.group",
    "process.env.LIVE_REPLAY_URL": "https://supergroup-cdn.mixin.group/live-replay/",
  },
})
