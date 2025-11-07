// CONTOH IMPLEMENTASI FRONTEND YANG ROBUST UNTUK ASSET CATEGORIES
// File: frontend/src/components/AssetCategoryManager.js

class AssetCategoryManager {
  constructor(apiBaseUrl = '/api/v1') {
    this.apiBaseUrl = apiBaseUrl;
    this.categories = [];
  }

  // Get auth token (sesuaikan dengan implementasi auth Anda)
  getAuthHeaders() {
    const token = localStorage.getItem('token');
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'Cache-Control': 'no-cache, no-store, must-revalidate', // Prevent caching
      'Pragma': 'no-cache',
      'Expires': '0'
    };
  }

  // Fetch categories with error handling
  async fetchCategories() {
    try {
      console.log('🔄 Fetching asset categories...');
      const response = await fetch(`${this.apiBaseUrl}/assets/categories`, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const result = await response.json();
      
      if (result.data && Array.isArray(result.data)) {
        this.categories = result.data;
        console.log(`✅ Loaded ${this.categories.length} categories`);
        this.updateUI();
        return this.categories;
      } else {
        console.error('❌ Invalid response format:', result);
        return [];
      }
    } catch (error) {
      console.error('❌ Error fetching categories:', error);
      alert('Failed to load categories: ' + error.message);
      return [];
    }
  }

  // Create category with robust error handling and state management
  async createCategory(categoryData) {
    try {
      console.log('🔄 Creating category:', categoryData);

      // Validate input
      if (!categoryData.code || !categoryData.name) {
        throw new Error('Code and name are required');
      }

      // Send create request
      const response = await fetch(`${this.apiBaseUrl}/assets/categories`, {
        method: 'POST',
        headers: this.getAuthHeaders(),
        body: JSON.stringify({
          code: categoryData.code.trim(),
          name: categoryData.name.trim(),
          description: categoryData.description?.trim() || '',
          is_active: categoryData.is_active !== false // default true
        })
      });

      const result = await response.json();

      if (!response.ok) {
        // Handle specific error cases
        if (result.details && result.details.includes('duplicate key')) {
          throw new Error('Category with this code or name already exists');
        }
        throw new Error(result.error || result.message || 'Unknown server error');
      }

      console.log('✅ Category created successfully:', result.data);

      // CRITICAL: Refresh the categories list immediately
      // Option 1: Add to local array (optimistic update)
      if (result.data) {
        this.categories.push(result.data);
        this.categories.sort((a, b) => a.name.localeCompare(b.name));
      }

      // Option 2: Fetch from server (more reliable)
      // Small delay to ensure database commit
      setTimeout(async () => {
        await this.fetchCategories();
      }, 100);

      return result.data;

    } catch (error) {
      console.error('❌ Error creating category:', error);
      alert('Failed to create category: ' + error.message);
      throw error;
    }
  }

  // Update UI with current categories
  updateUI() {
    const categoryListElement = document.getElementById('category-list');
    if (!categoryListElement) return;

    // Clear existing content
    categoryListElement.innerHTML = '';

    // Add categories to UI
    this.categories.forEach(category => {
      const categoryElement = document.createElement('div');
      categoryElement.className = 'category-item';
      categoryElement.innerHTML = `
        <div class="category-info">
          <strong>${category.name}</strong> (${category.code})
          <br>
          <small>${category.description || ''}</small>
        </div>
        <div class="category-actions">
          <button onclick="editCategory(${category.id})">Edit</button>
          <button onclick="deleteCategory(${category.id})">Delete</button>
        </div>
      `;
      categoryListElement.appendChild(categoryElement);
    });

    // Update category count
    const countElement = document.getElementById('category-count');
    if (countElement) {
      countElement.textContent = `${this.categories.length} categories`;
    }
  }

  // Handle form submission
  async handleFormSubmit(formElement) {
    const formData = new FormData(formElement);
    const categoryData = {
      code: formData.get('code'),
      name: formData.get('name'),
      description: formData.get('description'),
      is_active: formData.get('is_active') === 'on'
    };

    try {
      await this.createCategory(categoryData);
      
      // Reset form on success
      formElement.reset();
      
      // Show success message
      this.showMessage('Category created successfully!', 'success');
      
    } catch (error) {
      this.showMessage('Failed to create category: ' + error.message, 'error');
    }
  }

  // Show user feedback messages
  showMessage(message, type = 'info') {
    const messageElement = document.getElementById('message');
    if (messageElement) {
      messageElement.textContent = message;
      messageElement.className = `message ${type}`;
      messageElement.style.display = 'block';
      
      // Hide after 3 seconds
      setTimeout(() => {
        messageElement.style.display = 'none';
      }, 3000);
    }
  }

  // Initialize the category manager
  async init() {
    console.log('🚀 Initializing Asset Category Manager');
    
    // Load initial categories
    await this.fetchCategories();

    // Set up form handler if form exists
    const formElement = document.getElementById('category-form');
    if (formElement) {
      formElement.addEventListener('submit', async (e) => {
        e.preventDefault();
        await this.handleFormSubmit(formElement);
      });
    }

    // Set up refresh button
    const refreshButton = document.getElementById('refresh-categories');
    if (refreshButton) {
      refreshButton.addEventListener('click', () => {
        this.fetchCategories();
      });
    }
  }
}

// Usage example:
document.addEventListener('DOMContentLoaded', async () => {
  window.categoryManager = new AssetCategoryManager();
  await window.categoryManager.init();
});

// HTML Structure yang diperlukan:
/*
<div id="asset-category-manager">
  <div id="message" class="message" style="display: none;"></div>
  
  <form id="category-form">
    <input type="text" name="code" placeholder="Category Code" required>
    <input type="text" name="name" placeholder="Category Name" required>
    <textarea name="description" placeholder="Description (optional)"></textarea>
    <label>
      <input type="checkbox" name="is_active" checked> Active
    </label>
    <button type="submit">Add Category</button>
  </form>
  
  <div class="category-header">
    <h3>Categories <span id="category-count"></span></h3>
    <button id="refresh-categories">Refresh</button>
  </div>
  
  <div id="category-list"></div>
</div>
*/

// CSS untuk styling:
/*
.message {
  padding: 10px;
  margin: 10px 0;
  border-radius: 4px;
}
.message.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
.message.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
.message.info { background: #cce7ff; color: #004085; border: 1px solid #b8daff; }

.category-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  border: 1px solid #ddd;
  margin: 5px 0;
  border-radius: 4px;
}

.category-actions button {
  margin-left: 5px;
  padding: 5px 10px;
  border: none;
  border-radius: 3px;
  cursor: pointer;
}
*/