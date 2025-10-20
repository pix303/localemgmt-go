import { Component, computed, inject, OnInit } from '@angular/core';
import { UserStore } from '../../store/user/user-store';

@Component({
  selector: 'app-navbar',
  imports: [],
  templateUrl: './navbar.html',
  styleUrl: './navbar.css'
})
export class Navbar implements OnInit {

  readonly userStore = inject(UserStore);
  readonly userPicture = computed(() => {
    const user = this.userStore.user?.();
    return user?.picture ?? '/default-avatar.jpg';
  }
  );

  ngOnInit() {
  }
}
