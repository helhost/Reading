import auth from "../services/auth.js";
import { Toast } from "../components/Toast.js";

export async function getMe() {
  try {
    const me = await auth.me();
    if (me) return me;
  } catch (_) {
  }
  Toast("warn", "Unauthorized");
  location.hash = "#/login";
  return null;
}
