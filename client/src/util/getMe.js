import auth from "../services/auth.js";

export async function getMe() {
  try {
    const me = await auth.me();
    if (me) return me;
  } catch (_) {
  }
  location.hash = "#/login";
  return null;
}
