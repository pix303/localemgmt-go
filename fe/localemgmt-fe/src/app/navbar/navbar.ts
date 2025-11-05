import { Component, computed, inject, ChangeDetectionStrategy } from '@angular/core';
import { UserStore } from '../../store/user/user-store';

@Component({
  selector: 'app-navbar',
  imports: [],
  templateUrl: './navbar.html',
  styleUrl: './navbar.css',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class Navbar {
  readonly userStore = inject(UserStore);
  readonly userPicture = computed(() => {
    const user = this.userStore.user?.();
    return user?.picture ?? '/default-avatar.jpg';
  });

  logout() {
    console.log('logout call');
    this.userStore.logout();
  }
}
