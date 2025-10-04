import bookService from "./services/books.js";
import courseService from "./services/courses.js";

const books = await bookService.getAll();
const courses = await courseService.getAll(); // note to self, should change courses.getAll() to also give me all the books under that has it
console.log(courses, books)


