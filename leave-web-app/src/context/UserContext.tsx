import React, { createContext, useContext, useState } from 'react';
import { User, UserRole } from '../types';

interface UserContextValue {
  user: User | null;
  setUser: (user: User | null) => void;
}

const UserContext = createContext<UserContextValue>({
  user: null,
  setUser: () => {}
});

export function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  return (
    <UserContext.Provider value={{ user, setUser }}>
      {children}
    </UserContext.Provider>
  );
}

export function useUser() {
  return useContext(UserContext);
}

export const DEMO_USERS: { id: string; name: string; role: UserRole }[] = [
  { id: 'emp-001', name: 'Alice Johnson', role: 'employee' },
  { id: 'emp-002', name: 'Bob Smith', role: 'employee' },
  { id: 'mgr-001', name: 'Carol Williams', role: 'manager' }
];
