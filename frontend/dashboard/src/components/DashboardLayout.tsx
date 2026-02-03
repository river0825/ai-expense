'use client';

import React, { useState, useEffect } from 'react';
import { Sidebar } from './Sidebar';
import { TopBar } from './TopBar';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export function DashboardLayout({ children }: DashboardLayoutProps) {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [isMobile, setIsMobile] = useState(false);

  // Handle window resize to detect mobile
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 1024;
      setIsMobile(mobile);
      if (!mobile) {
        setIsSidebarOpen(false); // Reset mobile state when switching to desktop
      } else {
        setIsSidebarCollapsed(false); // Reset desktop collapse when switching to mobile
      }
    };

    // Initial check
    handleResize();

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  // Close sidebar when clicking outside on mobile
  const handleOverlayClick = () => {
    if (isMobile) {
      setIsSidebarOpen(false);
    }
  };

  return (
    <div className="min-h-screen bg-background font-sans text-text selection:bg-primary/30 relative">
      <Sidebar 
        isMobile={isMobile}
        isOpen={isSidebarOpen}
        isCollapsed={isSidebarCollapsed}
        onClose={() => setIsSidebarOpen(false)}
        onToggleCollapse={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
      />
      
      <TopBar 
        isMobile={isMobile}
        isSidebarCollapsed={isSidebarCollapsed}
        onMenuClick={() => setIsSidebarOpen(true)}
      />

      {/* Main Content Area */}
      <main 
        className={`
          transition-all duration-300 pt-20
          ${isMobile ? 'pl-0' : (isSidebarCollapsed ? 'pl-20' : 'pl-64')}
        `}
      >
        {children}

        {/* Mobile Overlay */}
        {isMobile && isSidebarOpen && (
          <div 
            className="fixed inset-0 bg-black/50 z-40 backdrop-blur-sm transition-opacity"
            onClick={handleOverlayClick}
          />
        )}
      </main>
    </div>
  );
}
