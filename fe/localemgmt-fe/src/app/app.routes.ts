import { Routes } from '@angular/router';
import { Home } from './home/home';
import { UserInfoComponent } from './userinfo/userinfo';
import { DebugComponent } from './debug/debug';

export const routes: Routes = [
  {
    path: 'home',
    component: Home,
  },
  {
    path: 'user-info',
    component: UserInfoComponent,
  },
  {
    path: 'debug',
    component: DebugComponent,
  },
  {
    path: '',
    redirectTo: 'home',
    pathMatch: 'full',
  },
];
