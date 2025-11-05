import { Component, computed, inject } from '@angular/core';
import { UserStore } from '../../store/user/user-store';
import { JsonPipe } from '@angular/common';

@Component({
  selector: 'app-debug',
  imports: [JsonPipe],
  templateUrl: './debug.html',
  styleUrl: './debug.css'
})
export class DebugComponent {
  readonly userStore = inject(UserStore);
  readonly userStoreState = computed(() => this.userStore.user?.() ?? {});
}
