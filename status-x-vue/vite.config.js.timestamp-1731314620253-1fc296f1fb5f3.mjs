// vite.config.js
import { defineConfig } from "file:///D:/learning/statusX/status-x-vue/node_modules/vite/dist/node/index.js";
import vue from "file:///D:/learning/statusX/status-x-vue/node_modules/@vitejs/plugin-vue/dist/index.mjs";
var vite_config_default = defineConfig({
  base: "/vue/",
  // 设置根路径
  plugins: [vue()],
  build: {
    outDir: "../monitor-server/frontend"
  },
  define: {
    "import.meta.env.MODE": JSON.stringify(process.env.NODE_ENV || "development")
  }
});
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcuanMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCJEOlxcXFxsZWFybmluZ1xcXFxzdGF0dXNYXFxcXHN0YXR1cy14LXZ1ZVwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRDpcXFxcbGVhcm5pbmdcXFxcc3RhdHVzWFxcXFxzdGF0dXMteC12dWVcXFxcdml0ZS5jb25maWcuanNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0Q6L2xlYXJuaW5nL3N0YXR1c1gvc3RhdHVzLXgtdnVlL3ZpdGUuY29uZmlnLmpzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZSdcbmltcG9ydCB2dWUgZnJvbSAnQHZpdGVqcy9wbHVnaW4tdnVlJ1xuXG4vLyBodHRwczovL3ZpdGUuZGV2L2NvbmZpZy9cbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XG4gIGJhc2U6ICcvdnVlLycsIC8vIFx1OEJCRVx1N0Y2RVx1NjgzOVx1OERFRlx1NUY4NFxuICBwbHVnaW5zOiBbdnVlKCldLFxuICBidWlsZDoge1xuICAgIG91dERpcjogJy4uL21vbml0b3Itc2VydmVyL2Zyb250ZW5kJyxcbiAgfSxcbiAgZGVmaW5lOiB7XG4gICAgJ2ltcG9ydC5tZXRhLmVudi5NT0RFJzogSlNPTi5zdHJpbmdpZnkocHJvY2Vzcy5lbnYuTk9ERV9FTlYgfHwgJ2RldmVsb3BtZW50JylcbiAgfVxufSlcbiJdLAogICJtYXBwaW5ncyI6ICI7QUFBMFIsU0FBUyxvQkFBb0I7QUFDdlQsT0FBTyxTQUFTO0FBR2hCLElBQU8sc0JBQVEsYUFBYTtBQUFBLEVBQzFCLE1BQU07QUFBQTtBQUFBLEVBQ04sU0FBUyxDQUFDLElBQUksQ0FBQztBQUFBLEVBQ2YsT0FBTztBQUFBLElBQ0wsUUFBUTtBQUFBLEVBQ1Y7QUFBQSxFQUNBLFFBQVE7QUFBQSxJQUNOLHdCQUF3QixLQUFLLFVBQVUsUUFBUSxJQUFJLFlBQVksYUFBYTtBQUFBLEVBQzlFO0FBQ0YsQ0FBQzsiLAogICJuYW1lcyI6IFtdCn0K
