import {WNodeHook} from '../../../../model/workflow.model';

export class HookEvent {
    type: string; // add update delete
    hook: WNodeHook;

    constructor(t: string, h: WNodeHook) {
        this.type = t;
        this.hook = h;
    }
}
