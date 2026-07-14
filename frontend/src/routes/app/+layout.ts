import { redirect } from "@sveltejs/kit";

export function load() {
  if (typeof localStorage !== "undefined") {
    const token = localStorage.getItem("token");
    if (!token) {
      redirect(302, "/auth/login");
    }
  }
}
