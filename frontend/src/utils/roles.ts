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
    case 'employee':
      return 'Employee';
    case 'purchasing':
      return 'Purchasing';
    case 'cost_control':
      return 'Cost Control';
    case 'gm':
      return 'GM';
    case 'project_director':
      return 'Direktur Proyek';
    case 'managing_director':
      return 'Direktur Utama';
    default:
      return humanizeRole(role);
  }
};

// Get name mapping for specific roles in Indonesian context
export const getRoleDisplayName = (role?: string | null): { role: string; example?: string } => {
  const r = normalizeRole(role);
  switch (r) {
    case 'employee':
      return { role: 'Employee', example: 'Andi (staff umum)' };
    case 'purchasing':
      return { role: 'Purchasing', example: 'Andi (Purchasing)' };
    case 'cost_control':
      return { role: 'Cost Control', example: 'Patrick' };
    case 'gm':
      return { role: 'GM', example: 'Marlin' };
    case 'project_director':
      return { role: 'Direktur Proyek', example: 'Christopher' };
    case 'managing_director':
      return { role: 'Direktur Utama', example: 'Jason' };
    case 'admin':
      return { role: 'Admin' };
    default:
      return { role: humanizeRole(role) };
  }
};

