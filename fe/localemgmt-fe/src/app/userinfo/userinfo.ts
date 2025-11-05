import { Component, computed, inject } from '@angular/core';
import { UserStore } from '../../store/user/user-store';

@Component({
  selector: 'app-userinfo',
  imports: [],
  templateUrl: './userinfo.html',
  styleUrl: './userinfo.css',
})
export class UserInfoComponent {
  readonly userStore = inject(UserStore);
  readonly user = computed(() => this.userStore.user?.());
  readonly contexts = computed(() => this.user()?.contexts ?? []);
  readonly username = computed(() => this.user()?.name ?? "");

  login() {
    this.userStore.login();
  }

  logout() {
    this.userStore.logout();
  }
}
