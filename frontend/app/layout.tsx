import type { Metadata } from "next";
import "./globals.css";
import ClientProviders from './ClientProviders';

export const metadata: Metadata = {
  title: "Unipro Project Manager - Cost Control Management System",
  description: "Professional project cost control and management system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700;800&display=swap" rel="stylesheet" />
        <script
          dangerouslySetInnerHTML={{
            __html: `
              // Global error handler to prevent external extension errors
              window.addEventListener('error', function(e) {
                if (e.error && e.error.message && e.error.message.includes('MetaMask')) {
                  e.preventDefault();
                  console.warn('MetaMask error suppressed:', e.error.message);
                  return false;
                }
              });
              
              // Theme initialization
              try {
                const theme = localStorage.getItem('theme');
                if (theme === 'dark' || (!theme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
                  document.documentElement.classList.add('dark');
                  document.documentElement.setAttribute('data-theme', 'dark');
                } else {
                  document.documentElement.classList.add('light');
                  document.documentElement.setAttribute('data-theme', 'light');
                }
              } catch (e) {
                console.warn('Theme initialization error:', e);
              }
              
              // Mark as hydrated to prevent FOUC
              window.addEventListener('DOMContentLoaded', function() {
                document.documentElement.classList.add('hydrated');
              });
              
              // Fallback if DOMContentLoaded already fired
              if (document.readyState === 'complete' || document.readyState === 'interactive') {
                document.documentElement.classList.add('hydrated');
              }
            `,
          }}
        />
      </head>
      <body>
        <ClientProviders>{children}</ClientProviders>
      </body>
    </html>
  );
}
