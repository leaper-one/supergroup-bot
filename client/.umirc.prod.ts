import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.RED_PACKET_URL": "https://red-api.mixinbots.com",
  },
})
