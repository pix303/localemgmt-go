import { Injectable, Signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { UserInfo } from '../store/user/user-store';
import { toSignal } from '@angular/core/rxjs-interop';
import { environment } from '../environments/environment';
import { Observable } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class UserService {

  constructor(private http: HttpClient, private router: Router) { }

  getUserInfo(): Observable<UserInfo> {
    const apiUrl = `${environment.api}/user/info`;
    const result = this.http.get<UserInfo>(apiUrl);
    return result;
  }

  login(): void {
    const apiUrl = `${environment.api}/login`;
    this.http.get(apiUrl)
      .subscribe({
        next: (res: any) => {
          window.location.href = res.url;
        },
        error: (err: any) => {
          console.error("Error logging in", err);
        }
      });
  }

  logout(): void {
    const apiUrl = `${environment.api}/user/logout`;
    this.http.post(apiUrl, {})
      .subscribe({
        next: () => {
          this.router.navigate(['/home']);
        },
        error: (err: any) => {
          console.error("Error logging out", err);
        }
      });
  }
}
