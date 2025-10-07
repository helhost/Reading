export default function Assignments(myCourses = [], remainingCourses = []) {
  const my = Array.isArray(myCourses) ? myCourses : [];
  const rem = Array.isArray(remainingCourses) ? remainingCourses : [];

  console.groupCollapsed("[Assignments] course lists");
  console.log("My courses (%d):", my.length, my);
  console.log("Remaining courses (%d):", rem.length, rem);
  console.groupEnd();
}
