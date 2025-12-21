<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Plus, Trash2, Package, Search, RefreshCw, Loader2 } from 'lucide-vue-next'
import { api, type RuntimeEnv, type RuntimePackage } from '@/api'
import { toast } from 'vue-sonner'

const availableRuntimes = ref<string[]>([])
const activeRuntime = ref('conda')
const envs = ref<RuntimeEnv[]>([])
const selectedEnv = ref<RuntimeEnv | null>(null)
const packages = ref<RuntimePackage[]>([])
const loading = ref(false)
const packagesLoading = ref(false)

// 创建环境
const showCreateDialog = ref(false)
const newEnvName = ref('')
const newEnvVersion = ref('')
const creating = ref(false)

// 删除环境
const showDeleteDialog = ref(false)
const envToDelete = ref<RuntimeEnv | null>(null)

// 安装包
const showInstallDialog = ref(false)
const packageToInstall = ref('')
const installing = ref(false)

// 搜索
const packageSearch = ref('')

const filteredPackages = computed(() => {
  if (!packageSearch.value) return packages.value
  const q = packageSearch.value.toLowerCase()
  return packages.value.filter(p => p.name.toLowerCase().includes(q))
})

async function loadRuntimes() {
  try {
    availableRuntimes.value = await api.runtime.getAvailable()
    if (availableRuntimes.value.length > 0 && !availableRuntimes.value.includes(activeRuntime.value)) {
      activeRuntime.value = availableRuntimes.value[0] ?? 'conda'
    }
  } catch {
    availableRuntimes.value = []
  }
}

async function loadEnvs() {
  if (!activeRuntime.value) return
  loading.value = true
  try {
    envs.value = await api.runtime.listEnvs(activeRuntime.value)
    if (envs.value.length > 0 && !selectedEnv.value) {
      const firstEnv = envs.value[0]
      if (firstEnv) selectEnv(firstEnv)
    }
  } catch {
    toast.error('加载环境列表失败')
    envs.value = []
  } finally {
    loading.value = false
  }
}

async function selectEnv(env: RuntimeEnv) {
  selectedEnv.value = env
  await loadPackages()
}

async function loadPackages() {
  if (!selectedEnv.value) return
  packagesLoading.value = true
  try {
    packages.value = await api.runtime.listPackages(activeRuntime.value, selectedEnv.value.name)
  } catch {
    toast.error('加载包列表失败')
    packages.value = []
  } finally {
    packagesLoading.value = false
  }
}

function openCreateDialog() {
  newEnvName.value = ''
  newEnvVersion.value = ''
  showCreateDialog.value = true
}

async function createEnv() {
  if (!newEnvName.value.trim()) {
    toast.error('请输入环境名称')
    return
  }
  creating.value = true
  try {
    await api.runtime.createEnv(activeRuntime.value, newEnvName.value.trim(), newEnvVersion.value.trim() || undefined)
    toast.success('环境创建成功')
    showCreateDialog.value = false
    await loadEnvs()
  } catch (e: unknown) {
    toast.error((e as Error).message || '创建失败')
  } finally {
    creating.value = false
  }
}

function confirmDeleteEnv(env: RuntimeEnv) {
  if (env.name === 'base') {
    toast.error('不能删除 base 环境')
    return
  }
  envToDelete.value = env
  showDeleteDialog.value = true
}

async function deleteEnv() {
  if (!envToDelete.value) return
  try {
    await api.runtime.deleteEnv(activeRuntime.value, envToDelete.value.name)
    toast.success('环境删除成功')
    if (selectedEnv.value?.name === envToDelete.value.name) {
      selectedEnv.value = null
      packages.value = []
    }
    await loadEnvs()
  } catch (e: unknown) {
    toast.error((e as Error).message || '删除失败')
  } finally {
    showDeleteDialog.value = false
    envToDelete.value = null
  }
}

function openInstallDialog() {
  packageToInstall.value = ''
  showInstallDialog.value = true
}

async function installPackage() {
  if (!packageToInstall.value.trim() || !selectedEnv.value) return
  installing.value = true
  try {
    await api.runtime.installPackage(activeRuntime.value, selectedEnv.value.name, packageToInstall.value.trim())
    toast.success('包安装成功')
    showInstallDialog.value = false
    await loadPackages()
  } catch (e: unknown) {
    toast.error((e as Error).message || '安装失败')
  } finally {
    installing.value = false
  }
}

async function uninstallPackage(pkg: RuntimePackage) {
  if (!selectedEnv.value) return
  try {
    await api.runtime.uninstallPackage(activeRuntime.value, selectedEnv.value.name, pkg.name)
    toast.success('包卸载成功')
    await loadPackages()
  } catch (e: unknown) {
    toast.error((e as Error).message || '卸载失败')
  }
}

function getRuntimeLabel(type: string) {
  const labels: Record<string, string> = { conda: 'Conda', node: 'Node.js' }
  return labels[type] || type
}

onMounted(async () => {
  await loadRuntimes()
  if (availableRuntimes.value.includes('conda')) {
    activeRuntime.value = 'conda'
    await loadEnvs()
  }
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">依赖管理</h2>
        <p class="text-muted-foreground">管理运行时环境和依赖包</p>
      </div>
    </div>

    <div v-if="availableRuntimes.length === 0" class="text-center py-8 text-muted-foreground">
      <Package class="h-10 w-10 mx-auto mb-3 opacity-50" />
      <p>未检测到可用的运行时环境</p>
      <p class="text-sm mt-1">请确保已安装 Conda 或其他支持的运行时</p>
    </div>

    <Tabs v-else v-model="activeRuntime" @update:model-value="loadEnvs">
      <TabsList>
        <TabsTrigger v-for="rt in availableRuntimes" :key="rt" :value="rt">
          {{ getRuntimeLabel(rt) }}
        </TabsTrigger>
      </TabsList>

      <TabsContent :value="activeRuntime" class="mt-4">
        <div class="flex gap-4 min-h-[480px]">
          <!-- 环境列表 -->
          <div class="w-52 shrink-0 border rounded-lg p-3 h-fit">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium">虚拟环境</span>
              <Button variant="ghost" size="icon" class="h-6 w-6" @click="openCreateDialog">
                <Plus class="h-3.5 w-3.5" />
              </Button>
            </div>
            <div class="space-y-0.5 min-h-[60px]">
              <div v-if="loading" class="text-sm text-muted-foreground text-center py-3">
                <Loader2 class="h-4 w-4 animate-spin mx-auto" />
              </div>
              <div
                v-else
                v-for="env in envs"
                :key="env.name"
                :class="[
                  'group flex items-center justify-between px-2 py-1.5 rounded cursor-pointer text-sm',
                  selectedEnv?.name === env.name ? 'bg-accent text-accent-foreground' : 'hover:bg-muted'
                ]"
                @click="selectEnv(env)"
              >
                <span class="truncate text-xs">{{ env.name }}</span>
                <Button
                  v-if="env.name !== 'base'"
                  variant="ghost"
                  size="icon"
                  class="h-5 w-5 shrink-0 opacity-0 group-hover:opacity-100"
                  @click.stop="confirmDeleteEnv(env)"
                >
                  <Trash2 class="h-3 w-3 text-destructive" />
                </Button>
              </div>
            </div>
          </div>

          <!-- 包列表 -->
          <div class="flex-1 border rounded-lg flex flex-col h-[480px]">
            <div class="flex items-center justify-between px-3 py-2 border-b bg-muted/30 shrink-0">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium">{{ selectedEnv?.name || '选择环境' }}</span>
                <Badge v-if="selectedEnv" variant="secondary" class="text-xs">{{ packages.length }} 包</Badge>
              </div>
              <div class="flex items-center gap-1.5">
                <div class="relative">
                  <Search class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground" />
                  <Input v-model="packageSearch" placeholder="搜索..." class="h-7 pl-7 w-36 text-xs" />
                </div>
                <Button variant="ghost" size="icon" class="h-7 w-7" @click="loadPackages" :disabled="!selectedEnv">
                  <RefreshCw class="h-3.5 w-3.5" />
                </Button>
                <Button size="sm" class="h-7 text-xs" @click="openInstallDialog" :disabled="!selectedEnv">
                  <Plus class="h-3.5 w-3.5 mr-1" /> 安装
                </Button>
              </div>
            </div>
            <div class="flex-1 overflow-y-auto">
              <div v-if="!selectedEnv" class="text-center py-8 text-muted-foreground text-sm">
                请选择一个环境
              </div>
              <div v-else-if="packagesLoading" class="text-center py-8">
                <Loader2 class="h-5 w-5 animate-spin mx-auto text-muted-foreground" />
              </div>
              <div v-else-if="filteredPackages.length === 0" class="text-center py-8 text-muted-foreground text-sm">
                {{ packageSearch ? '无匹配结果' : '暂无包' }}
              </div>
              <table v-else class="w-full text-sm">
                <thead class="bg-muted/50 sticky top-0">
                  <tr class="text-xs text-muted-foreground">
                    <th class="text-left px-3 py-1.5 font-medium">包名</th>
                    <th class="text-left px-3 py-1.5 font-medium">版本</th>
                    <th class="text-left px-3 py-1.5 font-medium">来源</th>
                    <th class="text-center px-3 py-1.5 font-medium w-16">操作</th>
                  </tr>
                </thead>
                <tbody class="divide-y">
                  <tr v-for="pkg in filteredPackages" :key="pkg.name" class="hover:bg-muted/50">
                    <td class="px-3 py-1.5 text-xs font-mono">{{ pkg.name }}</td>
                    <td class="px-3 py-1.5 text-xs text-muted-foreground">{{ pkg.version }}</td>
                    <td class="px-3 py-1.5 text-xs text-muted-foreground">{{ pkg.channel || '-' }}</td>
                    <td class="px-3 py-1.5 text-center">
                      <Button variant="ghost" size="icon" class="h-6 w-6 text-destructive" @click="uninstallPackage(pkg)">
                        <Trash2 class="h-3 w-3" />
                      </Button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </TabsContent>
    </Tabs>

    <!-- 创建环境对话框 -->
    <Dialog v-model:open="showCreateDialog">
      <DialogContent class="sm:max-w-[380px]">
        <DialogHeader>
          <DialogTitle>创建虚拟环境</DialogTitle>
        </DialogHeader>
        <div class="grid gap-3 py-3">
          <div class="grid grid-cols-4 items-center gap-3">
            <Label class="text-right text-sm">环境名称</Label>
            <Input v-model="newEnvName" placeholder="myenv" class="col-span-3 h-8" />
          </div>
          <div class="grid grid-cols-4 items-center gap-3">
            <Label class="text-right text-sm">Python</Label>
            <Input v-model="newEnvVersion" placeholder="3.10 (可选)" class="col-span-3 h-8" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" size="sm" @click="showCreateDialog = false">取消</Button>
          <Button size="sm" @click="createEnv" :disabled="creating">
            <Loader2 v-if="creating" class="h-3.5 w-3.5 mr-1.5 animate-spin" />
            创建
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除环境确认 -->
    <AlertDialog v-model:open="showDeleteDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>
            确定要删除环境 "{{ envToDelete?.name }}" 吗？此操作无法撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction class="bg-destructive text-white hover:bg-destructive/90" @click="deleteEnv">删除</AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 安装包对话框 -->
    <Dialog v-model:open="showInstallDialog">
      <DialogContent class="sm:max-w-[380px]">
        <DialogHeader>
          <DialogTitle>安装包</DialogTitle>
        </DialogHeader>
        <div class="grid gap-3 py-3">
          <div class="grid grid-cols-4 items-center gap-3">
            <Label class="text-right text-sm">包名</Label>
            <Input v-model="packageToInstall" placeholder="numpy" class="col-span-3 h-8" />
          </div>
          <p class="text-xs text-muted-foreground ml-auto col-span-4">
            支持版本指定: numpy==1.24.0
          </p>
        </div>
        <DialogFooter>
          <Button variant="outline" size="sm" @click="showInstallDialog = false">取消</Button>
          <Button size="sm" @click="installPackage" :disabled="installing">
            <Loader2 v-if="installing" class="h-3.5 w-3.5 mr-1.5 animate-spin" />
            安装
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
