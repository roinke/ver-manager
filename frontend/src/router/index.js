import { createRouter, createWebHashHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import BranchList from '../views/BranchList.vue'
import VersionList from '../views/VersionList.vue'
import ClientView from '../views/ClientView.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/branches', name: 'BranchList', component: BranchList },
  { path: '/versions', name: 'VersionList', component: VersionList },
  { path: '/timeline', name: 'ClientView', component: ClientView },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
