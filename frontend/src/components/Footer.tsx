import Link from "next/link";

export default function Footer() {
  return (
    <footer className="bg-gray-50 pt-16 pb-8 mt-auto border-t border-gray-100">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        
        {/* Trust Badges / Pre-footer Section */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-16 text-center">
          {/* Feature 1 */}
          <div className="flex flex-col items-center">
            <div className="w-12 h-12 bg-blue-100 text-blue-600 rounded-xl flex items-center justify-center mb-4">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>
            </div>
            <h4 className="font-bold text-gray-900 mb-2 text-sm">Instant Delivery</h4>
            <p className="text-xs text-gray-500 leading-relaxed max-w-[200px]">Automated systems deliver your credits in seconds.</p>
          </div>

          {/* Feature 2 */}
          <div className="flex flex-col items-center">
            <div className="w-12 h-12 bg-blue-100 text-blue-600 rounded-xl flex items-center justify-center mb-4">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>
            </div>
            <h4 className="font-bold text-gray-900 mb-2 text-sm">Secure Payments</h4>
            <p className="text-xs text-gray-500 leading-relaxed max-w-[200px]">All transactions are encrypted and 100% safe.</p>
          </div>

          {/* Feature 3 */}
          <div className="flex flex-col items-center">
            <div className="w-12 h-12 bg-blue-100 text-blue-600 rounded-xl flex items-center justify-center mb-4">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192l-3.536 3.536M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-5 0a4 4 0 11-8 0 4 4 0 018 0z" /></svg>
            </div>
            <h4 className="font-bold text-gray-900 mb-2 text-sm">24/7 Support</h4>
            <p className="text-xs text-gray-500 leading-relaxed max-w-[200px]">Our curator team is always ready to assist you.</p>
          </div>

          {/* Feature 4 */}
          <div className="flex flex-col items-center">
            <div className="w-12 h-12 bg-blue-100 text-blue-600 rounded-xl flex items-center justify-center mb-4">
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" /></svg>
            </div>
            <h4 className="font-bold text-gray-900 mb-2 text-sm">Best Pricing</h4>
            <p className="text-xs text-gray-500 leading-relaxed max-w-[200px]">Competitive rates for all your digital needs.</p>
          </div>
        </div>

        {/* Bottom Footer (Copyright & Links) */}
        <div className="flex flex-col md:flex-row justify-between items-center border-t border-gray-200 pt-8">
          <div className="mb-4 md:mb-0 text-center md:text-left">
            <h5 className="font-bold text-gray-800 text-sm mb-1">Markaz Digital Curator</h5>
            <p className="text-xs text-gray-500">© 2026 Markaz Digital Curator. All rights reserved.</p>
          </div>
          <div className="flex flex-wrap justify-center gap-6 text-xs font-medium text-gray-500">
            <Link href="/terms" className="hover:text-blue-700 hover:underline transition">Terms of Service</Link>
            <Link href="/privacy" className="hover:text-blue-700 hover:underline transition">Privacy Policy</Link>
            <Link href="/help" className="hover:text-blue-700 hover:underline transition">Help Center</Link>
            <Link href="/contact" className="hover:text-blue-700 hover:underline transition">Contact Us</Link>
            <Link href="/refund" className="hover:text-blue-700 hover:underline transition">Refund Policy</Link>
          </div>
        </div>

      </div>
    </footer>
  );
}