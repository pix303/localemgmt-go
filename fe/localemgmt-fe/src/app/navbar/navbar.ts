import { Component, inject } from '@angular/core';
import { UserStore } from '../../store/user/user-store';

@Component({
  selector: 'app-navbar',
  imports: [],
  templateUrl: './navbar.html',
  styleUrl: './navbar.css',
  providers: [UserStore],
})
export class Navbar {
  readonly userStore = inject(UserStore);
  test() {
    const userSignal = this.userStore.user();
    console.log(this.userStore.user().picture);
  }
}
