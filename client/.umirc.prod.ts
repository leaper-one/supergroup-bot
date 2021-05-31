import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.SERVER_URL": "",
    "process.env.CLIENT_ID": "",
    // 'process.env.ASSET_ID': 'c94ac88f-4671-3976-b60a-09064f1811e8',
    // 'process.env.AMOUNT': '0.01',

    "process.env.ASSET_ID": "965e5c6e-434c-3fa9-b780-c50f43cd955c",
    "process.env.AMOUNT": "1",
  },
})
