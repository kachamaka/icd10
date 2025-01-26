import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Icd10SearchService } from '../../services/icd10-search.service';
import { Icd10SearchResult } from '../../models/icd10-search-result';

@Component({
  selector: 'app-search-by-description',
  templateUrl: './search-by-description.component.html',
  styleUrls: ['./search-by-description.component.scss']
})
export class SearchByDescriptionComponent {
  searchForm: FormGroup;
  results: Icd10SearchResult[] = [];
  loading = false;

  constructor(
    private searchService: Icd10SearchService,
    private fb: FormBuilder
  ) {
    this.searchForm = this.fb.group({
      description: ['', Validators.required]
    });
  }

  get description() {
    return this.searchForm.get('description');
  }

  search() {
    if (this.searchForm.invalid) {
      return;
    }

    this.loading = true;
    this.results = [];

    this.searchService.searchByDescription(this.description?.value).subscribe(
      (data) => {
        console.log(data);
        if (data) {
          this.results = data;
        }
        this.loading = false;
      },
      (error) => {
        console.error('Error fetching results', error);
        this.loading = false;
      }
    );
  }
}
