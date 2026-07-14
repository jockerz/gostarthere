type AuthProvider = 'google' | 'github';

export interface UserAuthProvider {
  id: number;
  userId: number;
  provider: AuthProvider;
  providerUserId: string;
  email: string;
  createdAt: Date;
}
