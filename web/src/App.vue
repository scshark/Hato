<template>
    <n-config-provider :theme="theme" :theme-overrides="themeOverrides">
        <n-message-provider>
            <n-dialog-provider>
                <div
                    class="app-container"
                    :class="{ dark: theme?.name === 'dark' }"
                >
                    <div has-sider class="main-wrap" position="static">
                        <!-- 侧边栏 -->
                        <sidebar />

                        <div class="content-wrap">
                            <router-view
                                class="app-wrap"
                                v-slot="{ Component }"
                            >
                                <keep-alive>
                                    <component
                                        v-if="$route.meta.keepAlive"
                                        :is="Component"
                                    />
                                </keep-alive>
                                <component
                                    v-if="!$route.meta.keepAlive"
                                    :is="Component"
                                />
                            </router-view>
                        </div>

                        <!-- 右侧 -->
                        <rightbar />
                    </div>
                    <!-- 登录/注册公共组件 -->
                    <auth />
                </div>
            </n-dialog-provider>
        </n-message-provider>
        <n-global-style />
    </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';
import { darkTheme } from 'naive-ui';
import { version, buildTime } from '../build/info.json';

const store = useStore();
const theme = computed(() => (store.state.theme === 'dark' ? darkTheme : null));

/**
 * js 文件下使用这个做类型提示
 * @type import('naive-ui').GlobalThemeOverrides
 */
const themeOverrides = {
  common: {
    primaryColor:'#1da1f2',
    primaryColorHover:'#1a8cd8',
    primaryColorPressed:'#1576b6',
    primaryColorSuppl:'#1da1f2',
    fontSize:'15px',
  },
}
console.log(
    `%c Release Build Info 
%cVersion			v${version}
BuildTime		${buildTime}`,
    'background:#000;color:#FFF;font-weight:bold;',
    'background:#FFF;color:#000;'
);
</script>

<style lang="less">
@import '@/assets/css/main.less';
</style>