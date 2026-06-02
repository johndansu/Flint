import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Flint — The senior dev that never leaves',
  description:
    'A passive intelligence layer for your codebase. Catches what code review misses — before you commit.',
  icons: { icon: '/icon.svg', shortcut: '/icon.svg', apple: '/icon.svg' },
  openGraph: {
    title: 'Flint',
    description: 'The senior dev that never leaves.',
    url: 'https://flint.dev',
    siteName: 'Flint',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Flint — The senior dev that never leaves',
    description: 'A passive intelligence layer for your codebase.',
  },
};

// Runs synchronously before first paint — prevents flash of wrong theme.
const themeScript = `(function(){
  var s = localStorage.getItem('theme');
  var p = window.matchMedia('(prefers-color-scheme: dark)').matches;
  if (s === 'dark' || (!s && p) || (!s && !p)) {
    document.documentElement.classList.add('dark');
  }
})();`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* eslint-disable-next-line @next/next/no-before-interactive-script-component */}
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body>{children}</body>
    </html>
  );
}
