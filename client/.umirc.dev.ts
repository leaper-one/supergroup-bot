import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.RED_PACKET_URL": "http://192.168.2.153:8080",
  },
})
