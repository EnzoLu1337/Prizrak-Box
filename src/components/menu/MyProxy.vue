<script setup lang="ts">
import {useMenuStore} from "@/store/menuStore";
import {useI18n} from "vue-i18n";
import createApi from "@/api";
import {Events} from "@/runtime";
import {pError, pLoad, pSuccess, pWarning} from "@/util/pLoad";
import {useSettingStore} from "@/store/settingStore";
import {pUpdateMihomo} from "@/util/mihomo";
import {useHomeStore} from "@/store/homeStore";

// 使用store
const menuStore = useMenuStore();
const settingStore = useSettingStore();
const homeStore = useHomeStore();

// 获取当前 Vue 实例的 proxy 对象
const {proxy} = getCurrentInstance()!;
const api = createApi(proxy);

// 国际化
const {t} = useI18n();

// 页面使用参数
const tunOn = ref(false)


async function selected() {
  const list = await api.getProfileList()
  if (!list || list.length == 0) {
    pWarning(t("no-profile-warning"));
    return false
  }

  for (let profile of list) {
    if (profile['selected']) {
      return true
    }
  }

  pWarning(t("select-profile-warning"));
  return false
}


// 代理开关
async function doSwitch() {
  let ok = false

  // 检测通过执行后续操作
  if (!menuStore.proxy) {
    try {
      // 添加配置后执行
      const select = await selected()
      if (!select) {
        return
      }
      // 检测端口是否被占用
      await api.checkAddressPort({
        "bindAddress": settingStore.bindAddress,
        "port": settingStore.port,
      })
      // 开启代理
      await api.updateConfigs({
        "allow-lan": true,
        "mixed-port": settingStore.port,
        "bind-address": settingStore.bindAddress,
      })
      // 如果включен режим системного прокси, то включаем системный прокси
      if (settingStore.systemProxyMode) {
        await api.enableProxy({
          "bindAddress": settingStore.bindAddress,
          "port": settingStore.port,
        })
      }
      ok = true
      pSuccess(t("proxy-switch-on"));
    } catch (e) {
      if (e['message']) {
        pError(e['message'])
      }
    }
  } else {
    // Отключаем mixed-port в Mihomo, чтобы он перестал слушать порт
    await api.updateConfigs({
      "mixed-port": 0,
    })
    // Всегда отключаем системный прокси при выключении переключателя прокси
    await api.disableProxy()
    ok = true
    pWarning(t("proxy-switch-off"));
  }

  // 同步配置
  if (ok) {
    menuStore.setProxy(!menuStore.proxy);
    pUpdateMihomo(menuStore, settingStore, api)
  }

  // 发送事件通知
  Events.Emit({name: "proxy", data: menuStore.proxy});
}

const proxySwitch = async () => {
  await pLoad(t('switch.ing'), doSwitch)
}

Events.On("switchProxy", async () => {
  await proxySwitch()
});


// Диалог предложения установки сервиса
const showServiceDialog = ref(false);

// 虚拟网卡开关
async function tunSwitch() {
  // 检测是否运行在管理员模式下
  const admin = await api.getAdmin();
  if (!admin.data) {
    // Нет прав администратора - предлагаем установить сервис
    showServiceDialog.value = true;
    Events.Emit({name: "tun", data: false});
    return
  }

  // 添加配置后执行
  const select = await selected()
  if (!select) {
    Events.Emit({name: "tun", data: false});
    return
  }

  // Включаем TUN
  await enableTun();
}

// Включение TUN режима
async function enableTun() {
  menuStore.setTun(!tunOn.value);
  if (menuStore.tun) {
    api.updateConfigs({
      tun: {
        enable: true,
        stack: settingStore.stack,
      },
    }).then(() => {
      tunOn.value = true;
      pSuccess(t("tun-switch-on"));

      // 同步 mihomo 配置
      pUpdateMihomo(menuStore, settingStore, api)

      // 发送事件通知
      Events.Emit({name: "tun", data: menuStore.tun});
    });
  } else {
    api.updateConfigs({
      tun: {
        enable: false,
      },
    }).then(() => {
      tunOn.value = false;
      pWarning(t("tun-switch-off"));

      // 同步 mihomo 配置
      pUpdateMihomo(menuStore, settingStore, api)

      // 发送事件通知
      Events.Emit({name: "tun", data: menuStore.tun});
    });
  }
}

// Установка сервиса
async function installServiceHandler() {
  showServiceDialog.value = false;
  pLoad(t('service.installing'), async () => {
    try {
      // @ts-ignore
      const installed = await window.pxService.install();
      if (installed) {
        pSuccess(t('service.install-success'));
        // После установки сервиса нужен перезапуск приложения
        // чтобы backend запустился через сервис
        pWarning(t('service.restart-required'));
      } else {
        pError(t('service.install-failed'));
      }
    } catch (e) {
      pError(t('service.install-failed'));
    }
  });
}

// Закрыть диалог
function closeServiceDialog() {
  showServiceDialog.value = false;
}

Events.On("switchTun", async () => {
  await tunSwitch()
});


onMounted(async () => {
  if (homeStore.os != "Windows" && menuStore.tun) {
    await api.waitRunning()
    await tunSwitch()
  }
})

// Отслеживание изменения настройки "Режим системного прокси"
watch(() => settingStore.systemProxyMode, async (newValue, oldValue) => {
  // Применяем изменения только если прокси уже включен
  if (menuStore.proxy) {
    if (newValue) {
      // Включаем системный прокси
      try {
        await api.enableProxy({
          "bindAddress": settingStore.bindAddress,
          "port": settingStore.port,
        })
        pSuccess(t("proxy-switch-on"));
      } catch (e) {
        if (e['message']) {
          pError(e['message'])
        }
      }
    } else {
      // Выключаем системный прокси
      await api.disableProxy()
      pWarning(t("proxy-switch-off"));
    }
  }
})


</script>

<template>
  <div class="sub">
    <div class="switch-container">
      <span class="switch-label">
        {{ $t("proxy-switch") }}
      </span>
      <div
          :class="['switch', { 'switch-on': menuStore.proxy }]"
          @click="proxySwitch"
      >
        <div class="switch-circle"></div>
      </div>
    </div>
    <div class="switch-container">
      <span class="switch-label">
        {{ $t("tun-switch") }}
      </span>
      <div :class="['switch', { 'switch-on': tunOn }]" @click="tunSwitch">
        <div class="switch-circle"></div>
      </div>
    </div>
  </div>

  <!-- Диалог предложения установки сервиса -->
  <el-dialog
      v-model="showServiceDialog"
      :title="$t('service.dialog-title')"
      width="450px"
      :close-on-click-modal="true"
  >
    <div class="service-dialog">
      <p class="service-dialog__message">{{ $t('service.dialog-message') }}</p>
      <p class="service-dialog__description">{{ $t('service.dialog-description') }}</p>
    </div>
    <template #footer>
      <div class="service-dialog__footer">
        <el-button @click="closeServiceDialog">{{ $t('cancel') }}</el-button>
        <el-button type="primary" @click="installServiceHandler">{{ $t('service.install-btn') }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.sub {
  width: 184px; /* 卡片宽度 */
  background-color: var(--left-proxy-bg); /* 半透明白色背景 */
  border: 1px solid var(--sub-card-border);
  border-radius: 8px;
  box-shadow: var(--left-nav-shadow);
  line-height: 1.5;
  margin-left: 22px;
  margin-top: 25px;
  padding-bottom: 12px;
}

.sub:hover {
  box-shadow: var(--left-nav-hover-shadow);
}

.switch-container {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  padding-left: 10px;
  margin-top: 10px;
}

.switch-label {
  color: var(--text-color);
}

.switch {
  width: 54px;
  height: 26px;
  border: 2px solid var(--text-color);
  border-radius: 15px;
  position: relative;
  background-color: transparent;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.switch.switch-on {
  background-color: var(--left-item-selected-bg);
}

.switch-circle {
  width: 20px;
  height: 20px;
  background-color: var(--text-color);
  border-radius: 50%;
  position: absolute;
  top: 3px;
  left: 3px;
  transition: left 0.3s ease;
}

.switch-on .switch-circle {
  left: 31px;
}

/* Стили для диалога сервиса */
.service-dialog {
  padding: 10px 0;
}

.service-dialog__message {
  font-size: 16px;
  margin-bottom: 12px;
  color: var(--text-color);
}

.service-dialog__description {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.8;
  line-height: 1.6;
}

.service-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
