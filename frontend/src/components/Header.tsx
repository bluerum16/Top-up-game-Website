import Link from "next/link";

export default function Header() {
  return (
    <nav className="bg-white border-b border-gray-100 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-20">
          
          {/* Logo & Navigation Links */}
          <div className="flex items-center gap-10">
            <Link href="/" className="text-2xl font-bold text-blue-700">
              Markaz
            </Link>
            
            {/* Desktop Menu */}
            <div className="hidden md:flex space-x-8 h-full items-center">
              {/* Menu aktif ditandai dengan teks biru dan border bawah */}
              <Link href="/" className="text-blue-700 border-b-2 border-blue-700 h-20 flex items-center text-sm font-semibold">
                Home
              </Link>
              <Link href="/steam" className="text-gray-600 hover:text-blue-700 h-20 flex items-center text-sm font-medium transition-colors">
                Steam
              </Link>
              <Link href="/mlbb" className="text-gray-600 hover:text-blue-700 h-20 flex items-center text-sm font-medium transition-colors">
                Mobile Legends
              </Link>
              <Link href="/valorant" className="text-gray-600 hover:text-blue-700 h-20 flex items-center text-sm font-medium transition-colors">
                VALORANT
              </Link>
              <Link href="/pubgm" className="text-gray-600 hover:text-blue-700 h-20 flex items-center text-sm font-medium transition-colors">
                PUBG Mobile
              </Link>
            </div>
          </div>

          {/* Search Bar & Auth Buttons */}
          <div className="flex items-center gap-6">
            {/* Search Input (Hidden on mobile) */}
            <div className="hidden lg:flex items-center bg-gray-50 rounded-full px-4 py-2.5 border border-gray-200 focus-within:border-blue-300 focus-within:ring-2 focus-within:ring-blue-100 transition-all">
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input 
                type="text" 
                placeholder="Search games..." 
                className="bg-transparent border-none outline-none text-sm ml-2 w-48 text-gray-700 placeholder-gray-400"
              />
            </div>

            {/* Buttons */}
            <div className="flex items-center gap-4">
              <Link href="/login" className="text-gray-700 hover:text-blue-700 text-sm font-semibold px-2 transition-colors">
                Login
              </Link>
              <Link href="/register" className="bg-blue-700 hover:bg-blue-800 text-white text-sm font-semibold px-5 py-2.5 rounded-lg transition-colors shadow-sm">
                Register
              </Link>
            </div>
          </div>

        </div>
      </div>
    </nav>
  );
}