const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
})

function parseBirthday(birthday) {
  const [year, month, day] = birthday.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function formatBirthday(birthday) {
  return dateFormatter.format(parseBirthday(birthday))
}

export function calculateAge(birthday) {
  const birth = parseBirthday(birthday)
  const today = new Date()
  const birthdayIsUpcoming =
    today.getMonth() < birth.getMonth() ||
    (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())

  return today.getFullYear() - birth.getFullYear() - Number(birthdayIsUpcoming)
}
