import { inject } from '@angular/core';
import { patchState, signalStore, withHooks, withMethods, withProps, withState } from '@ngrx/signals';
import { UserService } from '../../services/user.service';

export type UserRole = 1 | 2 | 3;
export type UserInfo = {
  subjectId: string;
  name: string;
  email: string;
  picture: string;
  role: UserRole;
  contexts: string[];
};

type UserState = {
  isAuthenticated: boolean;
  user?: UserInfo;
};

const initialState: UserState = {
  isAuthenticated: false,
  user: undefined,
};

export const UserStore = signalStore(
  { providedIn: 'root' },
  withState(initialState),
  withProps(() => ({
    _userService: inject(UserService),
  })),
  withMethods((store) => ({

    loadUser() {
      store._userService.getUserInfo()
        .subscribe({
          next: (user) => {
            patchState(store, { user: user, isAuthenticated: true })
          },
          error: (err) => {
            console.log('error loading user', err)
            patchState(store, { user: undefined, isAuthenticated: false })
          }
        });
    },

    login() {
      if (!store.isAuthenticated()) {
        store._userService.login();
      } else {
        console.log('already logged in')
      }
    },

    logout() {
      if (store.isAuthenticated()) {
        store._userService.logout();
        patchState(store, { user: undefined, isAuthenticated: false });
      } else {
        console.log('already logged out')
      }
    },
  })),
  withHooks((store) => {
    return {
      onInit() {
        store.loadUser()
      }
    }
  })

);
