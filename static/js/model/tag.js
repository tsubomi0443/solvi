export class Tag {
  constructor({ name = '', count = 0 } = {}) {
    this.name = name;
    this.count = count;
  }

  static fromJSON(dto) {
    if (!dto) return new Tag();
    return new Tag({
      name: dto.name ?? '',
      count: Number(dto.count ?? 0),
    });
  }
}
