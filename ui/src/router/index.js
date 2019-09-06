import Vue from 'vue'
import Router from 'vue-router'
import CloudHome from '@/page/CloudHome'

Vue.use(Router)

export default new Router({
  routes: [{
    path: '/',
    name: 'CloudHome',
    component: CloudHome,
    children: [
      {
        path: '/key/Pipeline',
        name: 'Pipeline',
        component: () => import('@/page/Pipeline'),
      }
      
    ]
  }]
})
