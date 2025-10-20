import { Routes } from '@angular/router';
import { Home } from './home/home';
import { Userinfo } from './userinfo/userinfo';

export const routes: Routes = [

  {
    path: 'home',
    component: Home,
  },
  {
    path: 'user-info',
    component: Userinfo,
  },
  {
    path: '',
    redirectTo: 'home',
    pathMatch: 'full',
  },
];
