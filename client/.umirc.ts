import { defineConfig } from 'umi';
import routes from './src/routes';

// const VueCliPluginQiniuUploader = require("vue-cli-plugin-qiniu-uploader")

export default defineConfig({
  title: 'site.title',
  mfsu: {},
  dynamicImport: {
    loading: '@/components/Loading/index',
  },
  alias: {
    '@': require('path').resolve(__dirname, './src'),
  },
  nodeModulesTransform: {
    type: 'none',
  },
  routes,
  hash: true,
  fastRefresh: {},
  locale: {
    default: process.env.LANG === 'zh' ? 'zh' : 'en',
    title: true,
    antd: true,
  },
  extraBabelPlugins: ['babel-plugin-styled-components'],
  // chainWebpack: (config, a) => {
  //   if (process.env.NODE_ENV === "production") {
  //     const svgRule = config.module.rule("svg")
  //     svgRule.uses.clear()
  //     config.module
  //       .rule("images")
  //       .test(/\.(png|jpe?g|gif|webp|svg)(\?.*)?$/)
  //       .use("url-loader")
  //       .loader("url-loader")
  //       .tap((options: any) => Object.assign(options, { limit: 10240 }))
  //     return {
  //       // plugins: [new VueCliPluginQiniuUploader()],
  //     }
  //   }
  // },
  metas: [
    {
      name: 'theme-color',
      content: '#fff',
    },
  ],
});
