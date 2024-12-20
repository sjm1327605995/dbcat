import {createApp} from 'vue'
import App from './App.vue'

// Element Plus
import ElementPlus from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import 'element-plus/dist/index.css'

// Icons
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

// Global styles
import './style.css'

const app = createApp(App)

// 使用 Element Plus
app.use(ElementPlus, {
  locale: zhCn,
})

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.mount('#app')