export class SubBlockRegistry {
    map = new Map();
    register(r) {
        this.map.set(r.key, r);
    }
    async render(key, ctx) {
        const r = this.map.get(key);
        if (!r)
            throw new Error(`unknown sub-block: ${key}`);
        return r.render(ctx);
    }
    keys() {
        return Array.from(this.map.keys());
    }
}
