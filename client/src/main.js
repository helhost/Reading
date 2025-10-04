import bookService from "./services/books.js"

let a = await bookService.getAll()
console.log(a)


