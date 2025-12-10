import { UserRole } from '@/contexts/AuthContext';

/**
 * Get the default landing page for a user based on their role
 */
export function getDefaultPageForRole(role: UserRole | string): string {
    const normalizedRole = role.toLowerCase();

    switch (normalizedRole) {
        case 'admin':
        case 'managing_director':
            return '/dashboard';

        case 'purchasing':
            return '/dashboard';

        case 'cost_control':
            return '/cost-control';

        case 'gm':
        case 'project_director':
            return '/projects';

        case 'employee':
            return '/dashboard';

        default:
            return '/dashboard';
    }
}
