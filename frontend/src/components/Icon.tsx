import bellSvg from '../assets/icons/bell.svg?raw'
import userSvg from '../assets/icons/user.svg?raw'
import mailSvg from '../assets/icons/mail.svg?raw'
import mapSvg from '../assets/icons/map.svg?raw'
import plusSvg from '../assets/icons/plus.svg?raw'
import briefcaseSvg from '../assets/icons/briefcase.svg?raw'
import listSvg from '../assets/icons/list.svg?raw'
import zapSvg from '../assets/icons/zap.svg?raw'
import starSvg from '../assets/icons/star.svg?raw'
import clockSvg from '../assets/icons/clock.svg?raw'
import alertCircleSvg from '../assets/icons/alert-circle.svg?raw'
import checkSvg from '../assets/icons/check.svg?raw'
import logoutSvg from '../assets/icons/logout.svg?raw'
import userBigSvg from '../assets/icons/user-big.svg?raw'

const icons = {
  bell: bellSvg,
  user: userSvg,
  mail: mailSvg,
  map: mapSvg,
  plus: plusSvg,
  briefcase: briefcaseSvg,
  list: listSvg,
  zap: zapSvg,
  star: starSvg,
  clock: clockSvg,
  alertCircle: alertCircleSvg,
  check: checkSvg,
  logout: logoutSvg,
  userBig: userBigSvg,
}

export type IconName = keyof typeof icons

interface IconProps {
  name: IconName
  size?: number
}

export function Icon({ name, size = 20 }: IconProps) {
  return (
    <span
      className="icon"
      style={{ width: size, height: size }}
      dangerouslySetInnerHTML={{ __html: icons[name] }}
    />
  )
}
