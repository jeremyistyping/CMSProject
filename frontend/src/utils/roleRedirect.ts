import { UserRole } from '@/contexts/AuthContext';

/**
 * Get the default landing page for a user based on their role
 */
export function getDefaultPageForRole(role: UserRole | string): string {
    const normalizedRole = role.toLowerCase();

    switch (normalizedRole) {
        case 'admin':
        case 'director':
        case 'managing_director':
            return '/dashboard';

        case 'finance':
        case 'finance_manager':
            return '/dashboard';

        case 'purchasing':
            return '/cost-control/purchase-requests';

        case 'cost_control':
            return '/cost-control';

        case 'inventory_manager':
            return '/products';

        case 'gm':
        case 'project_director':
            return '/projects';

        case 'employee':
            return '/dashboard';

        default:
            return '/dashboard';
    }
}
