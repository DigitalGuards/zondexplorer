"use client";

export default function BlockLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="w-full min-h-screen">
      <div className="max-w-[1200px] mx-auto px-4">
        {children}
      </div>
    </div>
  );
}