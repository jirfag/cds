import { Component } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { finalize } from 'rxjs/internal/operators/finalize';
import { WorkflowTemplate } from '../../../../model/workflow-template.model';
import { WorkflowTemplateService } from '../../../../service/workflow-template/workflow-template.service';
import { Column } from '../../../../shared/table/data-table.component';

@Component({
    selector: 'app-workflow-template-list',
    templateUrl: './workflow-template.list.html',
    styleUrls: ['./workflow-template.list.scss']
})
export class WorkflowTemplateListComponent {
    loading: boolean;
    columns: Array<Column>;
    workflowTemplates: Array<WorkflowTemplate>;

    constructor(
        private _workflowTemplateService: WorkflowTemplateService,
        private _translate: TranslateService
    ) {
        this.columns = [
            <Column>{
                name: this._translate.instant('workflow_template_name'),
                selector: wt => wt.name
            },
            <Column>{
                name: this._translate.instant('workflow_template_description'),
                selector: wt => wt.description
            },
            <Column>{
                name: this._translate.instant('workflow_template_group'),
                selector: wt => wt.group.name
            }
        ];
        this.getTemplates();
    }

    getTemplates() {
        this.loading = true;
        this._workflowTemplateService.getWorkflowTemplates()
            .pipe(finalize(() => this.loading = false))
            .subscribe(wts => { this.workflowTemplates = wts; });
    }
}
