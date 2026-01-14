import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	// Enable standalone output for Docker deployment
	output: 'standalone',
	
	// Environment variables
	env: {
		NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
	},
	
	// TypeScript configuration
	typescript: {
		ignoreBuildErrors: true,
	},
	
	// ESLint configuration
	eslint: {
		ignoreDuringBuilds: true,
	},
	
	// Image optimization - disable for Docker to avoid issues
	images: {
		unoptimized: true,
	},
	
	// Disable telemetry
	telemetry: false,
	
	// Webpack configuration (only for non-Turbopack mode)
	...(!process.env.TURBOPACK && {
		webpack: (config, { isServer }) => {
			// Handle client-side module resolution for jsPDF
			if (!isServer) {
				config.resolve.fallback = {
					...config.resolve.fallback,
					fs: false,
					path: false,
				};
				
				// Handle jsPDF ES modules
				config.module.rules.push({
					test: /\.m?js$/,
					resolve: {
						fullySpecified: false
					}
				});
			}
			return config;
		},
	}),
	
	// API rewrites for development
	async rewrites() {
		// Use environment variable for API destination, fallback to localhost for development
		const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
		return [
			{
				source: '/api/:path*',
				destination: `${apiBaseUrl}/api/:path*`,
			},
		];
	},
};

export default nextConfig;
