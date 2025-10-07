import auth from "../services/auth.js";
import { toast } from "../components/toast.js";

export async function getMe() {
  try {
    const me = await auth.me();
    if (me) return me;
  } catch (_) {
  }
  toast("warn", "Unauthorized");
  location.hash = "#/login";
  return null;
}
