# 🔧 Frontend Modification: Add Project to Purchase Form

## 📋 Requirement
Berdasarkan meeting dengan client, setiap **Purchase Request** harus bisa di-link ke **Project** yang sedang berjalan untuk Cost Control tracking.

---

## 🎯 Changes Required

### 1. **Add Project Dropdown in Purchase Form**

#### **Location**: Purchase Form Modal (Basic Information Section)

**Before:**
```tsx
// Basic Information
- Vendor (dropdown)
- Purchase Date
- Due Date
- Discount
- Notes
```

**After:**
```tsx
// Basic Information
- Project (dropdown) ← NEW!
- Vendor (dropdown)
- Purchase Date
- Due Date
- Discount
- Notes
```

---

## 📝 Implementation Guide

### **Step 1: Create Project Service** (`src/services/projectService.ts`)

```typescript
import api from './api';

export interface Project {
  id: number;
  project_name: string;
  customer: string;
  city: string;
  project_type: string;
  budget: number;
  actual_cost: number;
  variance: number;
  status: string;
  overall_progress: number;
}

export const projectService = {
  // Get all active projects
  getActiveProjects: async (): Promise<Project[]> => {
    const response = await api.get('/projects', {
      params: { status: 'active' }
    });
    return response.data.data || response.data;
  },

  // Get project by ID
  getProjectById: async (id: number): Promise<Project> => {
    const response = await api.get(`/projects/${id}`);
    return response.data.data || response.data;
  },

  // Get project cost summary
  getProjectCostSummary: async (id: number) => {
    const response = await api.get(`/projects/${id}/cost-summary`);
    return response.data;
  }
};
```

---

### **Step 2: Update Purchase Form State**

Add `project_id` to form state:

```typescript
const [formData, setFormData] = useState({
  project_id: null, // ← NEW!
  vendor_id: '',
  date: new Date().toISOString().split('T')[0],
  due_date: '',
  discount: 0,
  ppn_rate: 11,
  payment_method: 'CREDIT',
  notes: '',
  items: []
});
```

---

### **Step 3: Fetch Projects in Form Component**

```typescript
import { projectService } from '@/services/projectService';

const PurchaseForm = () => {
  const [projects, setProjects] = useState([]);
  const [selectedProject, setSelectedProject] = useState(null);
  const [loadingProjects, setLoadingProjects] = useState(false);

  // Fetch active projects on mount
  useEffect(() => {
    fetchProjects();
  }, []);

  const fetchProjects = async () => {
    setLoadingProjects(true);
    try {
      const data = await projectService.getActiveProjects();
      setProjects(data);
    } catch (error) {
      console.error('Failed to fetch projects:', error);
      toast({
        title: 'Error',
        description: 'Failed to load projects',
        status: 'error',
        duration: 3000
      });
    } finally {
      setLoadingProjects(false);
    }
  };

  // Handle project selection
  const handleProjectChange = (e) => {
    const projectId = e.target.value ? parseInt(e.target.value) : null;
    setFormData({ ...formData, project_id: projectId });
    
    // Optionally: Load project details for budget warning
    if (projectId) {
      const project = projects.find(p => p.id === projectId);
      setSelectedProject(project);
    } else {
      setSelectedProject(null);
    }
  };

  // ... rest of component
};
```

---

### **Step 4: Add Project Field in JSX**

Insert this **BEFORE** the Vendor field in Basic Information section:

```tsx
<FormControl>
  <FormLabel>
    Project
    <Tooltip label="Link purchase to a project for cost tracking">
      <InfoIcon ml={2} boxSize={3} />
    </Tooltip>
  </FormLabel>
  <Select
    placeholder="Select project (optional)"
    value={formData.project_id || ''}
    onChange={handleProjectChange}
    isDisabled={loadingProjects}
  >
    {projects.map((project) => (
      <option key={project.id} value={project.id}>
        {project.project_name} - {project.city}
      </option>
    ))}
  </Select>
  <FormHelperText fontSize="xs" color="gray.500">
    Optional: Select project untuk tracking budget dan material cost
  </FormHelperText>
  
  {/* Show project budget info if selected */}
  {selectedProject && (
    <Alert status="info" mt={2} borderRadius="md" fontSize="sm">
      <AlertIcon />
      <Box>
        <Text fontWeight="medium">
          Budget: Rp {selectedProject.budget.toLocaleString('id-ID')}
        </Text>
        <Text fontSize="xs">
          Terpakai: Rp {selectedProject.actual_cost.toLocaleString('id-ID')} 
          ({((selectedProject.actual_cost / selectedProject.budget) * 100).toFixed(1)}%)
        </Text>
        <Text fontSize="xs" color={selectedProject.variance >= 0 ? 'green.600' : 'red.600'}>
          Sisa Budget: Rp {selectedProject.variance.toLocaleString('id-ID')}
        </Text>
      </Box>
    </Alert>
  )}
</FormControl>
```

---

### **Step 5: Update Submit Handler**

Make sure `project_id` is included in API payload:

```typescript
const handleSubmit = async () => {
  try {
    const payload = {
      project_id: formData.project_id, // ← Include this
      vendor_id: parseInt(formData.vendor_id),
      date: formData.date,
      due_date: formData.due_date,
      discount: parseFloat(formData.discount),
      ppn_rate: parseFloat(formData.ppn_rate),
      payment_method: formData.payment_method,
      notes: formData.notes,
      items: formData.items.map(item => ({
        product_id: item.product_id,
        quantity: parseInt(item.quantity),
        unit_price: parseFloat(item.unit_price),
        discount: parseFloat(item.discount || 0),
        expense_account_id: item.expense_account_id
      }))
    };

    await purchaseService.createPurchase(payload);
    
    toast({
      title: 'Success',
      description: 'Purchase created successfully',
      status: 'success'
    });
    
    onClose();
    refreshList();
  } catch (error) {
    toast({
      title: 'Error',
      description: error.response?.data?.message || 'Failed to create purchase',
      status: 'error'
    });
  }
};
```

---

### **Step 6: Update Purchase Table/List**

Add Project column in purchase list:

```tsx
<Td>{purchase.project?.project_name || '-'}</Td>
```

Full example:

```tsx
<Table>
  <Thead>
    <Tr>
      <Th>Code</Th>
      <Th>Project</Th> {/* ← NEW COLUMN */}
      <Th>Vendor</Th>
      <Th>Date</Th>
      <Th>Amount</Th>
      <Th>Status</Th>
      <Th>Actions</Th>
    </Tr>
  </Thead>
  <Tbody>
    {purchases.map((purchase) => (
      <Tr key={purchase.id}>
        <Td>{purchase.code}</Td>
        <Td>
          {purchase.project ? (
            <Badge colorScheme="purple">
              {purchase.project.project_name}
            </Badge>
          ) : (
            <Text color="gray.400">No Project</Text>
          )}
        </Td>
        <Td>{purchase.vendor.name}</Td>
        <Td>{formatDate(purchase.date)}</Td>
        <Td>Rp {purchase.total_amount.toLocaleString('id-ID')}</Td>
        <Td>
          <StatusBadge status={purchase.approval_status} />
        </Td>
        <Td>
          <ActionButtons purchase={purchase} />
        </Td>
      </Tr>
    ))}
  </Tbody>
</Table>
```

---

### **Step 7: Add Filter by Project**

In purchase list page, add project filter:

```tsx
const [filters, setFilters] = useState({
  project_id: '',
  status: '',
  vendor_id: '',
  start_date: '',
  end_date: ''
});

// In filter section
<FormControl w="200px">
  <FormLabel fontSize="sm">Project</FormLabel>
  <Select
    size="sm"
    value={filters.project_id}
    onChange={(e) => setFilters({ ...filters, project_id: e.target.value })}
  >
    <option value="">All Projects</option>
    {projects.map((project) => (
      <option key={project.id} value={project.id}>
        {project.project_name}
      </option>
    ))}
  </Select>
</FormControl>
```

---

## 🎨 UI/UX Enhancements

### **Budget Warning Alert**

When user selects a project with low remaining budget:

```tsx
{selectedProject && selectedProject.variance < 0 && (
  <Alert status="warning" mt={2}>
    <AlertIcon />
    <Box>
      <AlertTitle fontSize="sm">Project Over Budget!</AlertTitle>
      <AlertDescription fontSize="xs">
        Project "{selectedProject.project_name}" sudah melebihi budget
        sebesar Rp {Math.abs(selectedProject.variance).toLocaleString('id-ID')}
      </AlertDescription>
    </Box>
  </Alert>
)}

{selectedProject && selectedProject.variance > 0 && selectedProject.variance < (selectedProject.budget * 0.1) && (
  <Alert status="warning" mt={2}>
    <AlertIcon />
    <Box>
      <AlertTitle fontSize="sm">Budget Hampir Habis!</AlertTitle>
      <AlertDescription fontSize="xs">
        Sisa budget hanya Rp {selectedProject.variance.toLocaleString('id-ID')} 
        ({((selectedProject.variance / selectedProject.budget) * 100).toFixed(1)}%)
      </AlertDescription>
    </Box>
  </Alert>
)}
```

---

## 📊 Purchase Detail View Enhancement

When viewing purchase detail, show project info:

```tsx
{purchase.project && (
  <Box 
    p={4} 
    borderWidth="1px" 
    borderRadius="lg" 
    borderColor="purple.200"
    bg="purple.50"
  >
    <HStack justify="space-between" mb={2}>
      <Text fontWeight="bold" color="purple.700">
        📁 Project Information
      </Text>
      <Link href={`/projects/${purchase.project.id}`}>
        <Button size="xs" variant="ghost" colorScheme="purple">
          View Project Details →
        </Button>
      </Link>
    </HStack>
    
    <Grid templateColumns="repeat(2, 1fr)" gap={4}>
      <Box>
        <Text fontSize="xs" color="gray.600">Project Name</Text>
        <Text fontWeight="medium">{purchase.project.project_name}</Text>
      </Box>
      <Box>
        <Text fontSize="xs" color="gray.600">Location</Text>
        <Text fontWeight="medium">{purchase.project.city}</Text>
      </Box>
      <Box>
        <Text fontSize="xs" color="gray.600">Budget</Text>
        <Text fontWeight="medium">
          Rp {purchase.project.budget.toLocaleString('id-ID')}
        </Text>
      </Box>
      <Box>
        <Text fontSize="xs" color="gray.600">Budget Utilization</Text>
        <HStack>
          <Text fontWeight="medium">
            {((purchase.project.actual_cost / purchase.project.budget) * 100).toFixed(1)}%
          </Text>
          <Badge colorScheme={purchase.project.variance >= 0 ? 'green' : 'red'}>
            {purchase.project.variance >= 0 ? 'Under Budget' : 'Over Budget'}
          </Badge>
        </HStack>
      </Box>
    </Grid>
  </Box>
)}
```

---

## 🔍 Cost Control Dashboard Integration

Create a new component for Cost Control to see purchases by project:

```tsx
// src/components/cost-control/ProjectPurchaseSummary.tsx

const ProjectPurchaseSummary = ({ projectId }) => {
  const [purchases, setPurchases] = useState([]);
  const [summary, setSummary] = useState(null);

  useEffect(() => {
    fetchProjectPurchases();
  }, [projectId]);

  const fetchProjectPurchases = async () => {
    const response = await api.get('/purchases', {
      params: { project_id: projectId }
    });
    setPurchases(response.data.data);
    
    // Calculate summary
    const total = purchases.reduce((sum, p) => sum + p.total_amount, 0);
    const approved = purchases
      .filter(p => p.approval_status === 'APPROVED')
      .reduce((sum, p) => sum + p.total_amount, 0);
    
    setSummary({ total, approved });
  };

  return (
    <Box>
      <Heading size="md" mb={4}>Purchase Summary</Heading>
      <SimpleGrid columns={2} spacing={4} mb={4}>
        <Stat>
          <StatLabel>Total Purchases</StatLabel>
          <StatNumber>Rp {summary?.total.toLocaleString('id-ID')}</StatNumber>
          <StatHelpText>{purchases.length} transactions</StatHelpText>
        </Stat>
        <Stat>
          <StatLabel>Approved Purchases</StatLabel>
          <StatNumber>Rp {summary?.approved.toLocaleString('id-ID')}</StatNumber>
        </Stat>
      </SimpleGrid>
      
      <Table size="sm">
        <Thead>
          <Tr>
            <Th>Code</Th>
            <Th>Date</Th>
            <Th>Vendor</Th>
            <Th isNumeric>Amount</Th>
            <Th>Status</Th>
          </Tr>
        </Thead>
        <Tbody>
          {purchases.map((purchase) => (
            <Tr key={purchase.id}>
              <Td>{purchase.code}</Td>
              <Td>{formatDate(purchase.date)}</Td>
              <Td>{purchase.vendor.name}</Td>
              <Td isNumeric>Rp {purchase.total_amount.toLocaleString('id-ID')}</Td>
              <Td><StatusBadge status={purchase.approval_status} /></Td>
            </Tr>
          ))}
        </Tbody>
      </Table>
    </Box>
  );
};
```

---

## ✅ Checklist

- [ ] Create `projectService.ts`
- [ ] Add `project_id` to purchase form state
- [ ] Fetch active projects on form mount
- [ ] Add Project dropdown in Basic Information section
- [ ] Show project budget info when selected
- [ ] Add budget warning alerts
- [ ] Update submit handler to include `project_id`
- [ ] Add Project column in purchase list
- [ ] Add project filter in purchase list
- [ ] Update purchase detail view with project info
- [ ] Create ProjectPurchaseSummary component for Cost Control

---

## 🚀 Testing

1. Login as **Employee**
2. Create New Purchase
3. Select Project dari dropdown
4. Verify budget info ditampilkan
5. Submit purchase
6. Login as **Cost Control**
7. Verify purchase list menampilkan project name
8. Filter by project
9. Verify project cost tracking berfungsi

---

**Backend sudah ready!** Tinggal implement di frontend sesuai panduan ini. 🎯
