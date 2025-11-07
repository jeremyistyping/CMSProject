export const normalizeRole = (role?: string | null): string => {
  return (role || '').toString().trim().toLowerCase();
};

// Public alias for readability in UI code
export const toRoleKey = (role?: string | null): string => normalizeRole(role);

export const normalizeRoles = (roles?: (string | null | undefined)[]): string[] => {
  return (roles || []).map(r => normalizeRole(r)).filter(Boolean);
};

export const isRoleAllowed = (allowedRoles: (string | null | undefined)[] = [], role?: string | null): boolean => {
  const roleNorm = normalizeRole(role);
  const allowed = new Set(normalizeRoles(allowedRoles));
  return roleNorm !== '' && allowed.has(roleNorm);
};

export const humanizeRole = (role?: string | null): string => {
  const r = normalizeRole(role);
  if (!r) return 'Unknown';
  return r.charAt(0).toUpperCase() + r.slice(1);
};

// Format role for approval trail display
export const formatRoleForApproval = (role?: string | null): string => {
  const r = normalizeRole(role);
  switch (r) {
    case 'admin':
      return 'Admin';
    case 'director':
      return 'Director';
    case 'finance':
      return 'Finance Manager';
    case 'employee':
      return 'Employee';
    case 'inventory_manager':
      return 'Inventory Manager';
    default:
      return humanizeRole(role);
  }
};

// Get name mapping for specific roles in Indonesian context
export const getRoleDisplayName = (role?: string | null): { role: string; example?: string } => {
  const r = normalizeRole(role);
  switch (r) {
    case 'employee':
      return { role: 'Employee', example: 'John' };
    case 'finance':
      return { role: 'Finance Manager', example: 'Jack' };
    case 'director':
      return { role: 'Director', example: 'Josh' };
    case 'admin':
      return { role: 'Admin' };
    case 'inventory_manager':
      return { role: 'Inventory Manager' };
    default:
      return { role: humanizeRole(role) };
  }
};

