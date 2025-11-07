package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🔄 Updating Frontend to SSOT Endpoints")
	fmt.Println("======================================")
	fmt.Println("This will:")
	fmt.Println("1. Replace /journal-entries with /journals endpoints")
	fmt.Println("2. Update API service files")
	fmt.Println("3. Create SSOT service adapter")
	fmt.Println("")

	frontendDir := "../frontend/src"
	
	// Step 1: Update existing service files
	fmt.Println("1. 🔧 Updating existing service files...")
	if err := updateServiceFiles(frontendDir); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Step 2: Create SSOT service adapter
	fmt.Println("\n2. 📝 Creating SSOT service adapter...")
	if err := createSSOTServiceAdapter(frontendDir); err != nil {
		log.Printf("Error creating SSOT adapter: %v", err)
	}

	// Step 3: Update component files
	fmt.Println("\n3. 🔄 Updating component files...")
	if err := updateComponentFiles(frontendDir); err != nil {
		log.Printf("Warning: %v", err)
	}

	fmt.Println("\n🎉 Frontend Update Completed!")
	fmt.Println("============================")
	fmt.Println("✅ Service files updated")
	fmt.Println("✅ SSOT adapter created")
	fmt.Println("✅ Components updated")
	
	fmt.Println("\n💡 Next Steps:")
	fmt.Println("• Test frontend with new SSOT endpoints")
	fmt.Println("• Update any missing component imports")
	fmt.Println("• Verify all journal operations work")
}

func updateServiceFiles(frontendDir string) error {
	serviceDir := filepath.Join(frontendDir, "services")
	
	// Files to update
	filesToUpdate := []string{
		"journalIntegrationService.ts",
		"enhancedPLService.ts",
	}

	for _, fileName := range filesToUpdate {
		filePath := filepath.Join(serviceDir, fileName)
		
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("   ℹ️  File not found: %s\n", fileName)
			continue
		}

		// Read file
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to read %s: %v\n", fileName, err)
			continue
		}

		originalContent := string(content)
		
		// Create backup
		backupPath := filePath + ".backup"
		if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
			fmt.Printf("   ⚠️  Failed to create backup for %s: %v\n", fileName, err)
		}

		// Update endpoints
		updatedContent := updateEndpoints(originalContent)

		// Write updated content
		if err := ioutil.WriteFile(filePath, []byte(updatedContent), 0644); err != nil {
			fmt.Printf("   ⚠️  Failed to write updated %s: %v\n", fileName, err)
			continue
		}

		fmt.Printf("   ✅ Updated %s\n", fileName)
	}

	return nil
}

func updateEndpoints(content string) string {
	// Replace journal-entries endpoints with journals endpoints
	patterns := []string{
		// API endpoint replacements
		`/journal-entries`,
	}

	replacements := []string{
		"/journals",
	}

	updatedContent := content
	for i, pattern := range patterns {
		updatedContent = strings.ReplaceAll(updatedContent, pattern, replacements[i])
	}

	return updatedContent
}

func createSSOTServiceAdapter(frontendDir string) error {
	serviceDir := filepath.Join(frontendDir, "services")
	
	// Ensure services directory exists
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return fmt.Errorf("failed to create services directory: %v", err)
	}

	ssotServiceContent := `import { API_V1_BASE } from '@/config/api';

// SSOT Journal Entry Structure (aligned with backend)
export interface SSOTJournalEntry {
  id: number;
  entry_number: string;
  source_type: string;
  entry_date: string;
  description: string;
  reference?: string;
  notes?: string;
  total_debit: number;
  total_credit: number;
  status: 'DRAFT' | 'POSTED' | 'REVERSED';
  is_balanced: boolean;
  is_auto_generated: boolean;
  created_by: number;
  posted_by?: number;
  posted_at?: string;
  reversed_by?: number;
  reversed_at?: string;
  created_at: string;
  updated_at: string;
  journal_lines?: SSOTJournalLine[];
}

export interface SSOTJournalLine {
  id: number;
  journal_id: number;
  account_id: number;
  line_number: number;
  description: string;
  debit_amount: number;
  credit_amount: number;
  quantity?: number;
  unit_price?: number;
  created_at: string;
  updated_at: string;
  account?: {
    id: number;
    code: string;
    name: string;
    type: string;
  };
}

export interface CreateJournalRequest {
  entry_date: string;
  description: string;
  reference?: string;
  notes?: string;
  lines: Array<{
    account_id: number;
    description: string;
    debit_amount?: number;
    credit_amount?: number;
    quantity?: number;
    unit_price?: number;
  }>;
}

export interface UpdateJournalRequest extends CreateJournalRequest {
  // Same structure as create request
}

class SSOTJournalService {
  private getAuthHeaders() {
    const token = localStorage.getItem('token');
    return {
      'Authorization': ` + "`Bearer ${token}`" + `,
      'Content-Type': 'application/json',
    };
  }

  private buildQueryString(params: Record<string, any>): string {
    const searchParams = new URLSearchParams();
    
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '' && value !== 'ALL') {
        searchParams.append(key, value.toString());
      }
    });

    return searchParams.toString();
  }

  // Get all journal entries
  async getJournalEntries(params: {
    start_date?: string;
    end_date?: string;
    status?: string;
    source_type?: string;
    page?: number;
    limit?: number;
    search?: string;
  } = {}): Promise<{
    data: SSOTJournalEntry[];
    total: number;
    page: number;
    limit: number;
    totalPages: number;
  }> {
    const queryString = this.buildQueryString(params);
    const url = ` + "`${API_V1_BASE}/journals${queryString ? '?' + queryString : ''}`" + `;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(` + "`Failed to fetch journal entries: ${response.statusText}`" + `);
    }

    return await response.json();
  }

  // Get specific journal entry
  async getJournalEntry(id: number): Promise<SSOTJournalEntry> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/${id}`" + `, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error(` + "`Failed to fetch journal entry: ${response.statusText}`" + `);
    }

    const result = await response.json();
    return result.data;
  }

  // Create new journal entry
  async createJournalEntry(data: CreateJournalRequest): Promise<SSOTJournalEntry> {
    const response = await fetch(` + "`${API_V1_BASE}/journals`" + `, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to create journal entry');
    }

    const result = await response.json();
    return result.data;
  }

  // Update journal entry
  async updateJournalEntry(id: number, data: UpdateJournalRequest): Promise<SSOTJournalEntry> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/${id}`" + `, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to update journal entry');
    }

    const result = await response.json();
    return result.data;
  }

  // Delete journal entry
  async deleteJournalEntry(id: number): Promise<void> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/${id}`" + `, {
      method: 'DELETE',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to delete journal entry');
    }
  }

  // Post journal entry
  async postJournalEntry(id: number): Promise<SSOTJournalEntry> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/${id}/post`" + `, {
      method: 'POST',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to post journal entry');
    }

    const result = await response.json();
    return result.data;
  }

  // Reverse journal entry
  async reverseJournalEntry(id: number, reason?: string): Promise<SSOTJournalEntry> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/${id}/reverse`" + `, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify({ reason }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.message || 'Failed to reverse journal entry');
    }

    const result = await response.json();
    return result.data;
  }

  // Get journal summary
  async getJournalSummary(params: {
    start_date?: string;
    end_date?: string;
    status?: string;
  } = {}): Promise<{
    total_entries: number;
    total_debit: number;
    total_credit: number;
    posted_entries: number;
    draft_entries: number;
    reversed_entries: number;
  }> {
    const queryString = this.buildQueryString(params);
    const url = ` + "`${API_V1_BASE}/journals/summary${queryString ? '?' + queryString : ''}`" + `;
    
    const response = await fetch(url, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch journal summary');
    }

    const result = await response.json();
    return result.data;
  }

  // Get account balances
  async getAccountBalances(): Promise<Array<{
    account_id: number;
    account_code: string;
    account_name: string;
    debit_balance: number;
    credit_balance: number;
    balance: number;
    last_updated: string;
  }>> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/account-balances`" + `, {
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to fetch account balances');
    }

    const result = await response.json();
    return result.data;
  }

  // Refresh materialized view for account balances
  async refreshAccountBalances(): Promise<{ message: string; updated_at: string }> {
    const response = await fetch(` + "`${API_V1_BASE}/journals/account-balances/refresh`" + `, {
      method: 'POST',
      headers: this.getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error('Failed to refresh account balances');
    }

    const result = await response.json();
    return result.data;
  }

  // Legacy compatibility method - converts old journal entry calls to SSOT
  async getJournalEntriesLegacy(params: any): Promise<any> {
    console.warn('🔄 Using legacy compatibility method - please update to getJournalEntries()');
    
    // Map legacy parameters to SSOT parameters
    const ssotParams = {
      start_date: params.start_date,
      end_date: params.end_date,
      status: params.status || 'POSTED',
      page: params.page || 1,
      limit: params.limit || 100,
    };

    const result = await this.getJournalEntries(ssotParams);
    
    // Convert SSOT format to legacy format for backward compatibility
    return {
      data: result.data.map(entry => ({
        id: entry.id,
        code: entry.entry_number,
        description: entry.description,
        reference: entry.reference,
        reference_type: entry.source_type,
        entry_date: entry.entry_date,
        status: entry.status,
        total_debit: entry.total_debit,
        total_credit: entry.total_credit,
        is_balanced: entry.is_balanced,
        journal_lines: entry.journal_lines
      })),
      total: result.total
    };
  }
}

export const ssotJournalService = new SSOTJournalService();
export default ssotJournalService;
`

	filePath := filepath.Join(serviceDir, "ssotJournalService.ts")
	if err := ioutil.WriteFile(filePath, []byte(ssotServiceContent), 0644); err != nil {
		return fmt.Errorf("failed to write SSOT service: %v", err)
	}

	fmt.Printf("   ✅ Created ssotJournalService.ts\n")
	return nil
}

func updateComponentFiles(frontendDir string) error {
	// Find all TypeScript/JavaScript files that might need updating
	err := filepath.Walk(frontendDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") || strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx")) {
			// Skip service files (already updated)
			if strings.Contains(path, "services") {
				return nil
			}

			// Check if file contains journal-related imports/calls
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return nil // Skip files we can't read
			}

			contentStr := string(content)
			if strings.Contains(contentStr, "journal-entries") || strings.Contains(contentStr, "journalIntegrationService") {
				// Update this file
				if err := updateComponentFile(path, contentStr); err != nil {
					fmt.Printf("   ⚠️  Failed to update %s: %v\n", path, err)
				} else {
					relPath, _ := filepath.Rel(frontendDir, path)
					fmt.Printf("   ✅ Updated %s\n", relPath)
				}
			}
		}

		return nil
	})

	return err
}

func updateComponentFile(filePath, content string) error {
	// Create backup
	backupPath := filePath + ".backup"
	if err := ioutil.WriteFile(backupPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	updatedContent := content

	// Update imports
	updatedContent = strings.ReplaceAll(updatedContent, 
		"import { journalIntegrationService }", 
		"import { ssotJournalService }")
	
	updatedContent = strings.ReplaceAll(updatedContent, 
		"from '../services/journalIntegrationService'", 
		"from '../services/ssotJournalService'")
	
	updatedContent = strings.ReplaceAll(updatedContent, 
		"from '@/services/journalIntegrationService'", 
		"from '@/services/ssotJournalService'")

	// Update service calls
	updatedContent = strings.ReplaceAll(updatedContent, 
		"journalIntegrationService.getJournalEntries", 
		"ssotJournalService.getJournalEntriesLegacy")

	// Write updated content
	return ioutil.WriteFile(filePath, []byte(updatedContent), 0644)
}