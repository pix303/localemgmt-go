import { signalStore, withState } from '@ngrx/signals';

type UserRole = 1 | 2 | 3;
type UserInfo = {
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

export const UserStore = signalStore(withState(initialState));
