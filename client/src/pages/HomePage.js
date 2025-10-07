import getMe from "../util/index.js";

export default async function HomePage() {
  const me = await getMe();
  if (!me) return;

}
