<script setup lang="ts">
import {useI18n} from "vue-i18n";
import {pLoad, pSuccess, pError, pWarning} from "@/util/pLoad";

const {t} = useI18n();

// Состояние сервиса
const serviceStatus = ref<{
  installed: boolean;
  running: boolean;
  version?: string;
}>({
  installed: false,
  running: false
});

const loading = ref(false);

// Получение статуса сервиса
async function fetchServiceStatus() {
  try {
    // @ts-ignore
    const status = await window.pxService.getStatus();
    serviceStatus.value = status;
  } catch (e) {
    serviceStatus.value = {installed: false, running: false};
  }
}

// Установка сервиса
async function installService() {
  loading.value = true;
  try {
    // @ts-ignore
    const success = await window.pxService.install();
    if (success) {
      pSuccess(t('service.install-success'));
      pWarning(t('service.restart-required'));
      await fetchServiceStatus();
    } else {
      pError(t('service.install-failed'));
    }
  } catch (e) {
    pError(t('service.install-failed'));
  }
  loading.value = false;
}

// Удаление сервиса
async function uninstallService() {
  loading.value = true;
  try {
    // @ts-ignore
    const success = await window.pxService.uninstall();
    if (success) {
      pSuccess(t('service.uninstall-success'));
      pWarning(t('service.restart-required'));
      await fetchServiceStatus();
    } else {
      pError(t('service.uninstall-failed'));
    }
  } catch (e) {
    pError(t('service.uninstall-failed'));
  }
  loading.value = false;
}

// Статус в читаемом виде
const statusText = computed(() => {
  if (!serviceStatus.value.installed) {
    return t('service.status-not-installed');
  }
  if (serviceStatus.value.running) {
    return t('service.status-running');
  }
  return t('service.status-stopped');
});

const statusType = computed(() => {
  if (!serviceStatus.value.installed) {
    return 'info';
  }
  if (serviceStatus.value.running) {
    return 'success';
  }
  return 'warning';
});

// Проверяем статус при монтировании
onMounted(() => {
  fetchServiceStatus();
});
</script>

<template>
  <div class="service-setting">
    <div class="service-setting__header">
      <strong>{{ t('service.mode') }}:</strong>
      <el-tag :type="statusType" size="small" class="service-setting__status">
        {{ statusText }}
      </el-tag>
    </div>
    <p class="service-setting__description">{{ t('service.mode-description') }}</p>
    <div class="service-setting__actions">
      <el-button
          v-if="!serviceStatus.installed"
          type="primary"
          :loading="loading"
          @click="installService"
      >
        {{ t('service.install-btn') }}
      </el-button>
      <el-button
          v-else
          type="danger"
          :loading="loading"
          @click="uninstallService"
      >
        {{ t('service.uninstall-btn') }}
      </el-button>
      <el-button
          :loading="loading"
          @click="fetchServiceStatus"
      >
        {{ t('refresh') }}
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.service-setting {
  margin: 8px 0;
}

.service-setting__header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
}

.service-setting__status {
  margin-left: 8px;
}

.service-setting__description {
  font-size: 14px;
  color: var(--text-color);
  opacity: 0.7;
  margin: 8px 0;
}

.service-setting__actions {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}
</style>
