export class User {
  constructor({
    uuid = '',
    name = '',
    email = '',
    departmentName = '',
    icon = '',
    isSupporter = false,
    isAdmin = false,
  } = {}) {
    this.uuid = uuid;
    this.name = name;
    this.email = email;
    this.departmentName = departmentName;
    this.icon = icon;
    this.isSupporter = isSupporter;
    this.isAdmin = isAdmin;
  }

  static fromJSON(dto) {
    if (!dto) return new User();
    return new User({
      uuid: dto.uuid ?? '',
      name: dto.name ?? '',
      email: dto.email ?? '',
      departmentName: dto.departmentName ?? '',
      icon: dto.icon ?? '',
      isSupporter: Boolean(dto.isSupporter),
      isAdmin: Boolean(dto.isAdmin),
    });
  }
}
