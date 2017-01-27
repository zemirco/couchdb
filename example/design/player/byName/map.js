
function (doc) {
  if (doc.type === 'player') {
    emit(doc.name)
  }
}
