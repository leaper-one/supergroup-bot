import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.SERVER_URL": "http://192.168.2.153:7001",
    "process.env.CLIENT_ID": "",
    "process.env.RED_PACKET_URL": "http://192.168.2.153:8080",
    "process.env.ASSET_ID": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
    "process.env.AMOUNT": "1",
  },
})
