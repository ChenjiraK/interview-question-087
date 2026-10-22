const dateFormatter = new Intl.DateTimeFormat('en-GB', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
})

function parseBirthday(birthday: string): Date {
  const [year, month, day] = birthday.split('-').map(Number)
  return new Date(year, month - 1, day)
}

export function formatBirthday(birthday?: string | null): string {
  if (!birthday) return '-'
  return dateFormatter.format(parseBirthday(birthday))
}

export function calculateAge(birthday?: string | null): number | null {
  if (!birthday) return null
  const birth = parseBirthday(birthday)
  const today = new Date()
  const birthdayIsUpcoming =
    today.getMonth() < birth.getMonth() ||
    (today.getMonth() === birth.getMonth() && today.getDate() < birth.getDate())

  return today.getFullYear() - birth.getFullYear() - Number(birthdayIsUpcoming)
}
