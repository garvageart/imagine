export class DragData<T> {
    constructor(public readonly type: string, public readonly payload: T) {

    }

    toString(): string {
        return JSON.stringify({ type: this.type, payload: this.payload });
    }

    static fromString<T>(dataStr: string): DragData<T> | undefined {
        try {
            const obj = JSON.parse(dataStr);
            if (obj.type && obj.payload) {
                return new DragData<T>(obj.type, obj.payload);
            }
        } catch (e) {
            return undefined;
        }

        return undefined;
    }

    static fromJSON<T>(obj: any): DragData<T> | undefined {
        if (obj.type && obj.payload) {
            return new DragData<T>(obj.type, obj.payload);
        }

        return undefined;
    }

    toJSON(): any {
        return { type: this.type, payload: this.payload };
    }

    isInstanceOfType(type: DragData<T>) {
        return this.type === type.type;
    }

    setData(dataTransfer: DataTransfer) {
        dataTransfer.setData(this.type, this.toString());
    }

    static getData<T>(dataTransfer: DataTransfer, type: string): DragData<T> | undefined {
        const dataStr = dataTransfer.getData(type);
        if (dataStr) {
            return DragData.fromString<T>(dataStr);
        }

        return undefined;
    }

    static isType(dataTransfer: DataTransfer, type: string): boolean {
        return dataTransfer.types.includes(type);
    }
}