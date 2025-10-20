import { Component, computed, inject } from '@angular/core';
import { UserStore } from '../../store/user/user-store';

@Component({
  selector: 'app-userinfo',
  imports: [],
  templateUrl: './userinfo.html',
  styleUrl: './userinfo.css',
})
export class Userinfo {
  readonly userStore = inject(UserStore);
  readonly user = computed(() => this.userStore.user?.());
}
