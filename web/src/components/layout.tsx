import { Outlet } from "react-router-dom"
import { Footer } from "@/components/footer"

export function Layout() {
  return (
    <div className="flex min-h-screen flex-col">
      <main className="flex flex-1 flex-col">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
