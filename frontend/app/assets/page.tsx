'use client';

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { useTranslation } from '@/hooks/useTranslation';
import SimpleLayout from '@/components/layout/SimpleLayout';
import Table from '@/components/common/Table';
import {
  Box,
  Flex,
  Heading,
  Button,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
  useToast,
  FormControl,
  FormLabel,
  Input,
  Select,
  Textarea,
  Switch,
  VStack,
  HStack,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  Text,
  Badge,
  Icon,
  Image,
  Grid,
  GridItem,
  Tooltip,
} from '@chakra-ui/react';
import { FiPlus, FiEdit, FiTrash2, FiEye, FiDownload, FiMapPin, FiExternalLink, FiX, FiUpload, FiSettings, FiInfo } from 'react-icons/fi';
import { assetService, Asset as BackendAsset } from '@/services/assetService';
import { ASSET_CATEGORIES, DEPRECIATION_METHOD_LABELS, AssetsSummary } from '@/types/asset';
import AssetSummaryComponent from '@/components/assets/AssetSummary';
import InteractiveMapPicker from '@/components/common/InteractiveMapPicker';
import AssetImageUpload from '@/components/assets/AssetImageUpload';
import CurrencyInput from '@/components/common/CurrencyInput';
import { getAssetImageUrl } from '@/utils/imageUrl';
import { 
  validateAssetForm, 
  getFieldError, 
  hasFieldError, 
  ValidationError,
  AssetFormData as FormData
} from '@/utils/assetValidation';

// Use the form data interface from validation utils
type AssetFormData = FormData;

const AssetsPage = () => {
  const { token } = useAuth();
  const { t } = useTranslation();
  const [assets, setAssets] = useState<BackendAsset[]>([]);
  const [summary, setSummary] = useState<AssetsSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingSummary, setIsLoadingSummary] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const toast = useToast();
  
  // Tooltip descriptions for asset page
  const tooltips = {
    search: 'Cari aset berdasarkan nama, kode, serial number, atau lokasi',
    category: 'Kategori aset: Office Equipment (peralatan kantor), Vehicle (kendaraan), Building (bangunan), dll',
    status: 'Status aset: Active (digunakan), Disposed (dijual/dibuang), Under Maintenance (dalam perbaikan)',
    purchaseDate: 'Tanggal pembelian atau perolehan aset',
    purchasePrice: 'Harga perolehan aset (cost basis)',
    salvageValue: 'Nilai sisa/residu aset setelah masa manfaat habis',
    usefulLife: 'Masa manfaat aset dalam tahun (untuk perhitungan depresiasi)',
    depreciationMethod: 'Metode penyusutan: Straight Line (garis lurus), Declining Balance (saldo menurun)',
    condition: 'Kondisi fisik aset saat ini: Good (baik), Fair (cukup), Poor (buruk)',
    location: 'Lokasi fisik aset berada',
    serialNumber: 'Nomor seri atau identifikasi unik aset',
    assetAccount: 'Akun aset tetap di chart of accounts (opsional)',
    depreciationAccount: 'Akun beban penyusutan di chart of accounts (opsional)',
    notes: 'Catatan tambahan tentang aset (riwayat, spesifikasi, dll)',
  };

  // Filters & pagination
  const [searchTerm, setSearchTerm] = useState('');
  const [filterCategory, setFilterCategory] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState<BackendAsset | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>([]);
  const [isMapPickerOpen, setIsMapPickerOpen] = useState(false);
  
  // Image upload states
  const [pendingUpload, setPendingUpload] = useState<{assetId: number, file: File} | null>(null);
  const { isOpen: isAlertOpen, onOpen: onAlertOpen, onClose: onAlertClose } = useDisclosure();
  
  // Detail view states
  const [detailAsset, setDetailAsset] = useState<BackendAsset | null>(null);
  const { isOpen: isDetailOpen, onOpen: onDetailOpen, onClose: onDetailClose } = useDisclosure();
  
  // Category management states
  const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);

  // Recalculate depreciation (detail modal)
  const [recalcDate, setRecalcDate] = useState('');
  const [recalcResult, setRecalcResult] = useState<null | {
    current_book_value: number;
    accumulated_depreciation: number;
    depreciation_method: string;
  }>(null);
  const [isRecalcLoading, setIsRecalcLoading] = useState(false);
  // Database categories (with code) and merged name list for the dropdown
  const [dbCategories, setDbCategories] = useState<{ id: number; code: string; name: string; is_active: boolean }[]>([]);
  const [customCategories, setCustomCategories] = useState<string[]>([...ASSET_CATEGORIES]);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [editingCategoryIndex, setEditingCategoryIndex] = useState<number | null>(null);
  const [newCategoryCode, setNewCategoryCode] = useState('');
  
  // Account management states (only for optional fixed asset and depreciation accounts)
  const [fixedAssetAccounts, setFixedAssetAccounts] = useState<any[]>([]);
  const [depreciationAccounts, setDepreciationAccounts] = useState<any[]>([]);
  const [isLoadingAccounts, setIsLoadingAccounts] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState<AssetFormData>({
    name: '',
    category: '',
    status: 'ACTIVE',
    purchaseDate: '',
    purchasePrice: 0,
    salvageValue: 0,
    usefulLife: 1,
    depreciationMethod: 'STRAIGHT_LINE',
    isActive: true,
    notes: '',
    location: '',
    coordinates: '',
    serialNumber: '',
    condition: 'Good'
  });

  // Fetch assets from API
  const fetchAssets = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await assetService.getAssets();
      setAssets(response.data || []);
    } catch (error: any) {
      console.error('Error fetching assets:', error);
      setError(error.response?.data?.message || 'Failed to load assets. Please try again.');
      toast({
        title: 'Error',
        description: 'Failed to load assets',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setIsLoading(false);
    }
  };

  // Fetch asset summary
  const fetchAssetsSummary = async () => {
    try {
      setIsLoadingSummary(true);
      const response = await assetService.getAssetsSummary();
      setSummary(response.data);
    } catch (error: any) {
      console.error('Error fetching assets summary:', error);
      // Don't show error toast for summary as it's not critical
    } finally {
      setIsLoadingSummary(false);
    }
  };

  // Fetch accounts for dropdowns (only optional accounts for manual asset entry)
  const fetchAccounts = async () => {
    try {
      setIsLoadingAccounts(true);
      const [fixedAssetAccountsRes, depreciationAccountsRes] = await Promise.all([
        assetService.getFixedAssetAccounts(),
        assetService.getDepreciationExpenseAccounts()
      ]);
      
      setFixedAssetAccounts(fixedAssetAccountsRes.data || []);
      setDepreciationAccounts(depreciationAccountsRes.data || []);
    } catch (error: any) {
      console.error('Error fetching accounts:', error);
      // Don't show error toast for accounts as it's not critical
    } finally {
      setIsLoadingAccounts(false);
    }
  };

  // Fetch categories from database
  const fetchCategories = async () => {
    try {
      const response = await assetService.getAssetCategories();
      const cats = (response.data || []).map((c: any) => ({
        id: c.id,
        code: c.code || '',
        name: c.name,
        is_active: c.is_active !== false
      }));
      setDbCategories(cats);
      
      // Extract category names from response
      const categoryNames = cats.map((cat: any) => cat.name) || [];
      
      // Merge with default categories (avoid duplicates)
      const allCategories = [...new Set([...ASSET_CATEGORIES, ...categoryNames])];
      setCustomCategories(allCategories);
    } catch (error: any) {
      console.error('Error fetching categories:', error);
      // Fallback to default categories only
      setDbCategories([]);
      setCustomCategories([...ASSET_CATEGORIES]);
    }
  };

  // Load assets on component mount
  useEffect(() => {
    if (token) {
      fetchAssets();
      fetchAssetsSummary();
      fetchAccounts();
      fetchCategories(); // Load categories from database
    }
  }, [token]);

  // Handle form submission for create/update
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate form data (use dynamic categories loaded from DB)
    const errors = validateAssetForm(formData, customCategories);
    setValidationErrors(errors);
    
    if (errors.length > 0) {
      toast({
        title: 'Validation Error',
        description: `Please fix ${errors.length} error(s) in the form`,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
      return;
    }
    
    setIsSubmitting(true);
    setError(null);
    
    try {
      // Transform form data to match backend API
      const apiData = {
        // Note: code is auto-generated by backend if not provided
        name: formData.name,
        category: formData.category,
        status: formData.status,
        purchase_date: formData.purchaseDate.includes('T') ? formData.purchaseDate : `${formData.purchaseDate}T00:00:00Z`,
        purchase_price: formData.purchasePrice,
        salvage_value: formData.salvageValue || 0,
        useful_life: formData.usefulLife,
        depreciation_method: formData.depreciationMethod,
        is_active: formData.isActive !== false,
        notes: formData.notes || '',
        location: formData.location || '',
        coordinates: formData.coordinates || '',
        serial_number: formData.serialNumber || '',
        condition: formData.condition || 'Good',
        asset_account_id: formData.assetAccountId,
        depreciation_account_id: formData.depreciationAccountId,
        payment_method: formData.paymentMethod || 'CREDIT',
        payment_account_id: formData.paymentAccountId,
        credit_account_id: formData.creditAccountId,
        user_id: 1, // TODO: Get from auth context
      };
      
      if (formData.id) {
        await assetService.updateAsset(formData.id, apiData);
        toast({
          title: 'Success',
          description: 'Asset updated successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      } else {
        await assetService.createAsset(apiData);
        toast({
          title: 'Success',
          description: 'Asset created successfully',
          status: 'success',
          duration: 3000,
          isClosable: true,
        });
      }
      
      // Refresh assets list and summary
      await Promise.all([fetchAssets(), fetchAssetsSummary()]);
      
      // Close modal and reset form
      handleCloseModal();
    } catch (error: any) {
      let errorMsg = error.response?.data?.details || error.response?.data?.message || `Error ${formData.id ? 'updating' : 'creating'} asset`;
      
      // Handle specific duplicate code error
      if (errorMsg.includes('duplicate key value violates unique constraint') || 
          errorMsg.includes('uni_assets_code') ||
          errorMsg.includes('SQLSTATE 23505')) {
        errorMsg = 'An asset with this information already exists. Please check the asset name and category, then try again.';
      }
      
      setError(errorMsg);
      console.error('Error submitting asset:', error);
      
      toast({
        title: `Error ${formData.id ? 'Updating' : 'Creating'} Asset`,
        description: errorMsg,
        status: 'error',
        duration: 7000, // Longer duration for error messages
        isClosable: true,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle asset deletion
  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this asset?')) {
      return;
    }
    
    try {
      setError(null);
      await assetService.deleteAsset(id);
      toast({
        title: 'Success',
        description: 'Asset deleted successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
      // Refresh assets list and summary
      await Promise.all([fetchAssets(), fetchAssetsSummary()]);
    } catch (error: any) {
      const errorMsg = error.response?.data?.details || error.response?.data?.message || 'Failed to delete asset';
      setError(errorMsg);
      console.error('Error deleting asset:', error);
      toast({
        title: 'Error',
        description: errorMsg,
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Reset form data
  const resetForm = () => {
    setFormData({
      name: '',
      category: '',
      status: 'ACTIVE',
      purchaseDate: '',
      purchasePrice: 0,
      salvageValue: 0,
      usefulLife: 1,
      depreciationMethod: 'STRAIGHT_LINE',
      isActive: true,
      notes: '',
      location: '',
      coordinates: '',
      serialNumber: '',
      condition: 'Good'
    });
  };

  // Close modal and reset
  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedAsset(null);
    setError(null);
    setValidationErrors([]);
    resetForm();
  };

  // Open modal for creating a new asset
  const handleCreate = () => {
    setSelectedAsset(null);
    resetForm();
    setIsModalOpen(true);
  };

  // Open modal for editing an existing asset
  const handleEdit = (asset: BackendAsset) => {
    setSelectedAsset(asset);
    // Transform backend data to form format
    setFormData({
      id: asset.id,
      code: asset.code,
      name: asset.name,
      category: asset.category,
      status: asset.status as 'ACTIVE' | 'INACTIVE' | 'SOLD',
      purchaseDate: asset.purchase_date.split('T')[0], // Convert to YYYY-MM-DD
      purchasePrice: asset.purchase_price,
      salvageValue: asset.salvage_value,
      usefulLife: asset.useful_life,
      depreciationMethod: asset.depreciation_method as 'STRAIGHT_LINE' | 'DECLINING_BALANCE',
      isActive: asset.is_active,
      notes: asset.notes,
      location: asset.location || '',
      coordinates: asset.coordinates || '',
      serialNumber: asset.serial_number || '',
      condition: asset.condition || 'Good',
      assetAccountId: asset.asset_account_id,
      depreciationAccountId: asset.depreciation_account_id
    });
    setIsModalOpen(true);
  };

  // Handle form input changes
  // Suggest accounts based on category name
  const suggestAccountsForCategory = (categoryName: string) => {
    if (!categoryName) return { assetId: formData.assetAccountId, depId: formData.depreciationAccountId };
    const name = categoryName.toLowerCase();
    let fa = fixedAssetAccounts.find(acc => acc.code?.startsWith('1502') || /bangunan|building/.test(acc.name?.toLowerCase() || ''))
          || fixedAssetAccounts.find(acc => acc.code?.startsWith('1503') && /vehicle|kendaraan/.test(name))
          || fixedAssetAccounts.find(acc => acc.code?.startsWith('1504') && /furniture/.test(name))
          || fixedAssetAccounts.find(acc => acc.code?.startsWith('1505') && /(computer|it)/.test(name))
          || fixedAssetAccounts.find(acc => acc.code?.startsWith('1501') && /(machinery|mesin)/.test(name))
          || fixedAssetAccounts[0];
    let dep = depreciationAccounts.find(acc => /depreciation|penyusutan/.test((acc.name||'').toLowerCase())) || depreciationAccounts[0];
    return { assetId: fa?.id, depId: dep?.id };
  };

  const handleInputChange = (field: keyof AssetFormData, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));

    // Auto-suggest accounts when category changes and no account selected yet
    if (field === 'category') {
      const { assetId, depId } = suggestAccountsForCategory(value as string);
      setFormData(prev => ({
        ...prev,
        assetAccountId: prev.assetAccountId || assetId,
        depreciationAccountId: prev.depreciationAccountId || depId,
      }));
    }
  };

  // Format currency for display
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  // Calculate current book value
  const calculateBookValue = (asset: BackendAsset) => {
    return asset.purchase_price - asset.accumulated_depreciation;
  };

  // Get status badge color
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'ACTIVE':
        return 'green';
      case 'INACTIVE':
        return 'gray';
      case 'SOLD':
        return 'red';
      default:
        return 'gray';
    }
  };


  // Table columns definition
  const columns = [
    { 
      header: 'Code', 
      accessor: (asset: BackendAsset) => (
        <Text 
          fontWeight="medium" 
          color="blue.600"
          fontSize="sm"
          whiteSpace="nowrap"
        >
          {asset.code}
        </Text>
      )
    },
    { 
      header: 'Name', 
      accessor: (asset: BackendAsset) => (
        <Text fontSize="sm" fontWeight="medium" noOfLines={1}>
          {asset.name}
        </Text>
      )
    },
    { 
      header: 'Category', 
      accessor: (asset: BackendAsset) => (
        <Text fontSize="sm" noOfLines={1}>
          {asset.category}
        </Text>
      )
    },
    { 
      header: 'Purchase Price', 
      accessor: (asset: BackendAsset) => (
        <Text fontSize="sm" fontWeight="medium" whiteSpace="nowrap">
          {formatCurrency(asset.purchase_price)}
        </Text>
      )
    },
    { 
      header: 'Book Value', 
      accessor: (asset: BackendAsset) => (
        <Text fontSize="sm" fontWeight="medium" whiteSpace="nowrap">
          {formatCurrency(calculateBookValue(asset))}
        </Text>
      )
    },
    { 
      header: 'Status', 
      accessor: (asset: BackendAsset) => (
        <Badge colorScheme={getStatusColor(asset.status)} variant="subtle" size="sm">
          {asset.status}
        </Badge>
      )
    },
    { 
      header: 'Location', 
      accessor: (asset: BackendAsset) => (
        <VStack align="start" spacing={1} maxW="180px">
          <Text noOfLines={1} fontSize="xs">
            {asset.location || 'No location'}
          </Text>
          {asset.coordinates && (
            <HStack spacing={1}>
              <Text fontSize="xs" color="gray.500" fontFamily="mono" noOfLines={1}>
                {asset.coordinates}
              </Text>
              <Button
                size="xs"
                variant="ghost"
                colorScheme="blue"
                onClick={() => assetService.openInMaps(asset.coordinates!)}
                title="View on Maps"
                minW="auto"
                p={1}
              >
                <FiMapPin size={10} />
              </Button>
            </HStack>
          )}
        </VStack>
      )
    },
  ];

  // Handle location picker
  const handleLocationPick = (locationData: { name: string; description: string; address: string; coordinates: string }) => {
    // Update coordinates field
    handleInputChange('coordinates', locationData.coordinates);
    
    // Update location field with comprehensive info
    let locationText = locationData.name;
    if (locationData.description) {
      locationText += ` - ${locationData.description}`;
    }
    if (locationData.address) {
      locationText += ` (${locationData.address})`;
    }
    handleInputChange('location', locationText);
  };

  // Action buttons for each row
  const renderActions = (asset: BackendAsset) => (
    <>
      <Button
        size="xs"
        variant="outline"
        leftIcon={<FiEye />}
        onClick={() => handleViewDetails(asset)}
        colorScheme="blue"
        minW="auto"
        px={2}
      >
        View
      </Button>
      <Button
        size="xs"
        variant="outline"
        leftIcon={<FiEdit />}
        onClick={() => handleEdit(asset)}
        minW="auto"
        px={2}
      >
        Edit
      </Button>
      <Button
        size="xs"
        colorScheme="red"
        variant="outline"
        leftIcon={<FiTrash2 />}
        onClick={() => handleDelete(asset.id)}
        minW="auto"
        px={2}
      >
        Delete
      </Button>
      <Input
        type="file"
        accept="image/*"
        onChange={(e) => handleFileChange(e, asset.id)}
        style={{ display: 'none' }}
        id={`file-upload-${asset.id}`}
      />
      <Button
        size="xs"
        variant="outline"
        leftIcon={<FiUpload />}
        as="label"
        htmlFor={`file-upload-${asset.id}`}
        cursor="pointer"
        minW="auto"
        px={2}
        whiteSpace="nowrap"
      >
        {asset.image_path ? 'Update' : 'Upload'}
      </Button>
    </>
  );

  // Handle view asset details
  const handleViewDetails = async (asset: BackendAsset) => {
    setRecalcResult(null);
    setRecalcDate('');
    try {
      const response = await assetService.getAsset(asset.id);
      setDetailAsset(response.data);
      onDetailOpen();
    } catch (error: any) {
      console.error('Error fetching asset details:', error);
      toast({
        title: 'Error',
        description: 'Failed to load asset details',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };


  // Recalculate depreciation as of a date for selected detail asset
  const handleRecalculate = async () => {
    if (!detailAsset) return;
    try {
      setIsRecalcLoading(true);
      const res = await assetService.calculateDepreciation(detailAsset.id, recalcDate || undefined);
      setRecalcResult({
        current_book_value: res.data.current_book_value,
        accumulated_depreciation: res.data.accumulated_depreciation,
        depreciation_method: res.data.depreciation_method,
      });
      toast({ title: 'Recalculated', status: 'success', duration: 2000, isClosable: true });
    } catch (e) {
      toast({ title: 'Failed to recalculate', status: 'error', duration: 3000, isClosable: true });
    } finally {
      setIsRecalcLoading(false);
    }
  };

  // Handle export assets
  const handleExport = () => {
    if (assets.length === 0) {
      toast({
        title: 'No Data',
        description: 'No assets to export',
        status: 'warning',
        duration: 3000,
        isClosable: true,
      });
      return;
    }
    
    assetService.exportToCSV(assets);
    toast({
      title: 'Export Started',
      description: 'Assets data is being downloaded',
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
  };

  // Handle file change for image upload
  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>, assetId: number) => {
    if (e.target.files && e.target.files[0]) {
      const file = e.target.files[0];
      const asset = assets.find(a => a.id === assetId);
      
      if (asset && asset.image_path) {
        // Asset already has an image, show confirmation
        setPendingUpload({ assetId, file });
        onAlertOpen();
      } else {
        // No existing image, upload directly
        handleUpload(assetId, file);
      }
    }
  };

  // Confirm image update
  const confirmImageUpdate = () => {
    if (pendingUpload) {
      handleUpload(pendingUpload.assetId, pendingUpload.file);
      setPendingUpload(null);
    }
    onAlertClose();
  };

  // Handle direct upload
  const handleUpload = async (assetId: number, file: File) => {
    try {
      const response = await assetService.uploadAssetImage(assetId, file);
      
      // Update the asset in the list with the new image path
      setAssets(prevAssets => 
        prevAssets.map(asset => 
          asset.id === assetId 
            ? { ...asset, image_path: response.path }
            : asset
        )
      );
      
      toast({
        title: 'Image uploaded successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      
      // Reset file input
      const fileInput = document.getElementById(`file-upload-${assetId}`) as HTMLInputElement;
      if (fileInput) {
        fileInput.value = '';
      }
    } catch (error) {
      toast({
        title: 'Failed to upload image',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Handle image upload from component
  const handleImageUpload = async (updatedAsset: BackendAsset) => {
    // Update the selectedAsset with new image path
    setSelectedAsset(updatedAsset);
    
    // Update the assets list with the updated asset
    setAssets(prevAssets => 
      prevAssets.map(asset => 
        asset.id === updatedAsset.id ? updatedAsset : asset
      )
    );
    
    toast({
      title: 'Image Updated',
      description: 'Asset image has been updated successfully',
      status: 'success',
      duration: 3000,
      isClosable: true,
    });
  };

  // Category management functions
  const handleOpenCategoryModal = () => {
    setIsCategoryModalOpen(true);
    setNewCategoryName('');
    setNewCategoryCode('');
    setEditingCategoryIndex(null);
  };

  const handleCloseCategoryModal = () => {
    setIsCategoryModalOpen(false);
    setNewCategoryName('');
    setNewCategoryCode('');
    setEditingCategoryIndex(null);
  };

  const handleAddCategory = async () => {
    if (!newCategoryName.trim()) {
      toast({
        title: 'Error',
        description: 'Category name cannot be empty',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (!newCategoryCode.trim()) {
      toast({
        title: 'Error',
        description: 'Category code/prefix cannot be empty',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    const code = newCategoryCode.trim().toUpperCase().replace(/[^A-Z0-9-]/g, '').substring(0, 10);
    if (code.length < 2) {
      toast({
        title: 'Error',
        description: 'Category code must be at least 2 characters',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (customCategories.includes(newCategoryName.trim())) {
      toast({
        title: 'Error',
        description: 'Category already exists',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    try {
      // Call backend API to create asset category
      const categoryData = {
        code,
        name: newCategoryName.trim(),
        description: '',
        is_active: true
      };
      
      await assetService.createAssetCategory(categoryData as any);
      
      // Refresh categories from database to get the latest list
      await fetchCategories();
      setNewCategoryName('');
      setNewCategoryCode('');
      
      toast({
        title: 'Success',
        description: 'Category added and saved to database successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    } catch (error: any) {
      console.error('Error creating category:', error);
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to create category',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    }
  };

  // Helper to get code for a given category name (from DB or default mapping)
  const getCodeForCategory = (name: string): string => {
    const fromDb = dbCategories.find(c => c.name === name)?.code;
    if (fromDb) return fromDb;
    // Default mapping (same as backend static mapping)
    const map: Record<string, string> = {
      'Fixed Asset': 'FA',
      'Real Estate': 'RE',
      'Computer Equipment': 'CE',
      'Vehicle': 'VH',
      'Office Equipment': 'OE',
      'Furniture': 'FR',
      'IT Infrastructure': 'IT',
      'Machinery': 'MC',
      'Tanah': 'LAND'
    };
    return map[name] || '';
  };

  const handleEditCategory = (index: number) => {
    setEditingCategoryIndex(index);
    const name = customCategories[index];
    setNewCategoryName(name);
    setNewCategoryCode(getCodeForCategory(name));
  };

  const handleUpdateCategory = () => {
    if (!newCategoryName.trim()) {
      toast({
        title: 'Error',
        description: 'Category name cannot be empty',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    if (editingCategoryIndex !== null) {
      const updatedCategories = [...customCategories];
      updatedCategories[editingCategoryIndex] = newCategoryName.trim();
      setCustomCategories(updatedCategories);
      setEditingCategoryIndex(null);
      setNewCategoryName('');
      toast({
        title: 'Success',
        description: 'Category updated successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  const handleDeleteCategory = (index: number) => {
    const categoryToDelete = customCategories[index];
    
    // Check if category is being used by any assets
    const isUsed = assets.some(asset => asset.category === categoryToDelete);
    if (isUsed) {
      toast({
        title: 'Cannot Delete Category',
        description: 'This category is being used by one or more assets. Please reassign those assets first.',
        status: 'warning',
        duration: 5000,
        isClosable: true,
      });
      return;
    }

    if (window.confirm(`Are you sure you want to delete the category "${categoryToDelete}"?`)) {
      const updatedCategories = customCategories.filter((_, i) => i !== index);
      setCustomCategories(updatedCategories);
      toast({
        title: 'Success',
        description: 'Category deleted successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
    }
  };

  const cancelEdit = () => {
    setEditingCategoryIndex(null);
    setNewCategoryName('');
  };

  return (
<SimpleLayout allowedRoles={['admin', 'finance', 'director', 'employee']}>
      <Box>
        <Flex justify="space-between" align="center" mb={6}>
          <Box>
            <Heading size="lg">Asset Master</Heading>
            <Text color="gray.600" mt={1}>
              Manage company assets and depreciation
            </Text>
          </Box>
          <Flex gap={3}>
            <Button
              variant="outline"
              leftIcon={<FiDownload />}
              onClick={handleExport}
              isDisabled={assets.length === 0}
            >
              Export
            </Button>
            <Button
              variant="outline"
              leftIcon={<FiSettings />}
              onClick={handleOpenCategoryModal}
              colorScheme="gray"
            >
              Manage Categories
            </Button>
            <Button
              colorScheme="blue" 
              leftIcon={<FiPlus />}
              onClick={handleCreate}
            >
              {t('assets.addAsset')}
            </Button>
          </Flex>
        </Flex>
        
        {error && (
          <Alert status="error" mb={4}>
            <AlertIcon />
            <AlertTitle mr={2}>Error!</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        
        <AssetSummaryComponent 
          summary={summary} 
          isLoading={isLoadingSummary} 
        />
        
        {/* Filters */}
        <Box mb={4}>
          <HStack spacing={3} align="center" wrap="wrap">
            <Input
              placeholder={t('assets.searchAssets')}
              value={searchTerm}
              onChange={(e) => { setSearchTerm(e.target.value); setCurrentPage(1); }}
              maxW="280px"
            />
            <Select
              placeholder="All categories"
              value={filterCategory}
              onChange={(e) => { setFilterCategory(e.target.value); setCurrentPage(1); }}
              maxW="240px"
            >
              {Array.from(new Set(assets.map(a => a.category).filter(Boolean))).map(cat => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </Select>
            <Select
              placeholder="All status"
              value={filterStatus}
              onChange={(e) => { setFilterStatus(e.target.value); setCurrentPage(1); }}
              maxW="200px"
            >
              <option value="ACTIVE">ACTIVE</option>
              <option value="INACTIVE">INACTIVE</option>
              <option value="SOLD">SOLD</option>
            </Select>
            <HStack spacing={2}>
              <Text fontSize="sm" color="gray.600">Rows:</Text>
              <Select value={pageSize} onChange={(e) => { setPageSize(parseInt(e.target.value) || 10); setCurrentPage(1); }} maxW="80px">
                <option value={5}>5</option>
                <option value={10}>10</option>
                <option value={20}>20</option>
                <option value={50}>50</option>
              </Select>
            </HStack>
          </HStack>
        </Box>
        
        {/* Filter & paginate data */}
        {(() => {
          const filtered = assets.filter(a => {
            const matchSearch = !searchTerm ||
              a.code?.toLowerCase().includes(searchTerm.toLowerCase()) ||
              a.name?.toLowerCase().includes(searchTerm.toLowerCase());
            const matchCat = !filterCategory || a.category === filterCategory;
            const matchStatus = !filterStatus || a.status === filterStatus;
            return matchSearch && matchCat && matchStatus;
          });
          const total = filtered.length;
          const totalPages = Math.max(1, Math.ceil(total / pageSize));
          const page = Math.min(currentPage, totalPages);
          const start = (page - 1) * pageSize;
          const pageData = filtered.slice(start, start + pageSize);
          
          return (
            <>
              <Table<BackendAsset>
                columns={columns}
                data={pageData}
                keyField="id"
                title={`Assets (${total})`}
                actions={renderActions}
                isLoading={isLoading}
                emptyMessage="No assets found. Click 'Add Asset' to create your first asset."
              />
              {/* Pagination Controls */}
              <HStack justify="space-between" mt={3}>
                <Text fontSize="sm" color="gray.600">Showing {Math.min(total, start + 1)}–{Math.min(total, start + pageSize)} of {total}</Text>
                <HStack>
                  <Button size="sm" onClick={() => setCurrentPage(1)} isDisabled={page<=1}>First</Button>
                  <Button size="sm" onClick={() => setCurrentPage(p => Math.max(1, p-1))} isDisabled={page<=1}>Prev</Button>
                  <Text fontSize="sm">Page {page} / {totalPages}</Text>
                  <Button size="sm" onClick={() => setCurrentPage(p => Math.min(totalPages, p+1))} isDisabled={page>=totalPages}>Next</Button>
                  <Button size="sm" onClick={() => setCurrentPage(totalPages)} isDisabled={page>=totalPages}>Last</Button>
                </HStack>
              </HStack>
            </>
          );
        })()}
        
        <Modal 
          isOpen={isModalOpen} 
          onClose={handleCloseModal} 
          size="6xl"
        >
          <ModalOverlay />
          <ModalContent>
            <form onSubmit={handleSubmit}>
              <ModalHeader>
                {selectedAsset?.id ? t('assets.editAsset') : t('assets.addAsset')}
              </ModalHeader>
              <ModalCloseButton />
              
              <ModalBody pb={6}>
                {/* Information Banner */}
                <Alert status="info" borderRadius="md" bg="blue.50" border="1px solid" borderColor="blue.200">
                  <AlertIcon color="blue.500" />
                  <Box>
                    <AlertTitle color="blue.700" fontSize="sm" fontWeight="bold">
                      📝 Manual Asset Entry
                    </AlertTitle>
                    <AlertDescription color="blue.600" fontSize="xs" mt={1}>
                      This form is for recording existing assets only (no financial transactions). 
                      For new asset purchases, use the Purchases module to create purchase orders.
                    </AlertDescription>
                  </Box>
                </Alert>
                
                <VStack spacing={6}>
                  {/* Basic Information Section */}
                  <Box w="full">
                    <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                      📋 Basic Information
                    </Text>
                    <VStack spacing={4}>
                      <HStack w="full" spacing={4}>
                        <FormControl isRequired isInvalid={hasFieldError(validationErrors, 'name')}>
                          <FormLabel fontSize="sm" fontWeight="medium">Asset Name</FormLabel>
                          <Input
                            value={formData.name || ''}
                            onChange={(e) => handleInputChange('name', e.target.value)}
                            placeholder="Enter asset name"
                            size="md"
                          />
                          {getFieldError(validationErrors, 'name') && (
                            <Text fontSize="xs" color="red.500" mt={1}>
                              {getFieldError(validationErrors, 'name')}
                            </Text>
                          )}
                        </FormControl>
                        
                        <FormControl isRequired isInvalid={hasFieldError(validationErrors, 'category')}>
                          <FormLabel fontSize="sm" fontWeight="medium">Category</FormLabel>
                          <Select
                            value={formData.category || ''}
                            onChange={(e) => handleInputChange('category', e.target.value)}
                            placeholder="Select category"
                            size="md"
                          >
                            {customCategories.map((category) => (
                              <option key={category} value={category}>
                                {category}
                              </option>
                            ))}
                          </Select>
                          {getFieldError(validationErrors, 'category') && (
                            <Text fontSize="xs" color="red.500" mt={1}>
                              {getFieldError(validationErrors, 'category')}
                            </Text>
                          )}
                        </FormControl>
                      </HStack>
                      
                      <HStack w="full" spacing={4}>
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">Serial Number</FormLabel>
                          <Input
                            value={formData.serialNumber || ''}
                            onChange={(e) => handleInputChange('serialNumber', e.target.value)}
                            placeholder="Enter serial number"
                            size="md"
                          />
                        </FormControl>
                        
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">Condition</FormLabel>
                          <Select
                            value={formData.condition || 'Good'}
                            onChange={(e) => handleInputChange('condition', e.target.value)}
                            size="md"
                          >
                            <option value="Excellent">Excellent</option>
                            <option value="Good">Good</option>
                            <option value="Fair">Fair</option>
                            <option value="Poor">Poor</option>
                          </Select>
                        </FormControl>
                        
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">Status</FormLabel>
                          <HStack spacing={4} pt={2}>
                            <Switch
                              isChecked={formData.isActive !== false}
                              onChange={(e) => handleInputChange('isActive', e.target.checked)}
                              colorScheme="green"
                            />
                            <Text fontSize="sm" color={formData.isActive ? 'green.600' : 'red.500'}>
                              {formData.isActive ? 'Active' : 'Inactive'}
                            </Text>
                          </HStack>
                        </FormControl>
                      </HStack>
                    </VStack>
                  </Box>

                  {/* Financial Information Section */}
                  <Box w="full">
                    <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                      💰 Financial Information
                    </Text>
                    <VStack spacing={4}>
                      <HStack w="full" spacing={4}>
                        <FormControl isRequired isInvalid={hasFieldError(validationErrors, 'purchaseDate')}>
                          <FormLabel fontSize="sm" fontWeight="medium">Purchase Date</FormLabel>
                          <Input
                            type="date"
                            value={formData.purchaseDate || ''}
                            onChange={(e) => handleInputChange('purchaseDate', e.target.value)}
                            size="md"
                          />
                          {getFieldError(validationErrors, 'purchaseDate') && (
                            <Text fontSize="xs" color="red.500" mt={1}>
                              {getFieldError(validationErrors, 'purchaseDate')}
                            </Text>
                          )}
                        </FormControl>
                        
                        <CurrencyInput
                          value={formData.purchasePrice || 0}
                          onChange={(value) => handleInputChange('purchasePrice', value)}
                          label="Purchase Price"
                          placeholder="Contoh: Rp 100.000.000"
                          isRequired={true}
                          isInvalid={hasFieldError(validationErrors, 'purchasePrice')}
                          errorMessage={getFieldError(validationErrors, 'purchasePrice')}
                          size="md"
                          min={1}
                        />
                      </HStack>
                      
                      <HStack w="full" spacing={4}>
                        <FormControl isInvalid={hasFieldError(validationErrors, 'salvageValue')}>
                          <HStack spacing={2} align="center">
                            <FormLabel fontSize="sm" fontWeight="medium" mb={0}>Salvage Value</FormLabel>
                            <Tooltip
                              label={
                                <Box>
                                  <Text fontWeight="semibold" mb={1}>💡 Salvage Value (Nilai Sisa)</Text>
                                  <Text fontSize="xs" lineHeight="1.4">
                                    Perkiraan nilai asset di akhir masa manfaat.
                                  </Text>
                                  <Text fontSize="xs" lineHeight="1.4" mt={1}>
                                    Contoh: Mobil Rp 100 juta, setelah 5 tahun masih 
                                    bernilai Rp 20 juta = salvage value Rp 20 juta.
                                  </Text>
                                  <Text fontSize="xs" lineHeight="1.4" mt={1} fontWeight="medium">
                                    Mempengaruhi perhitungan depresiasi bulanan.
                                  </Text>
                                </Box>
                              }
                              hasArrow
                              placement="top"
                              bg="gray.800"
                              color="white"
                              borderRadius="md"
                              p={3}
                              maxW="280px"
                            >
                              <span><Icon as={FiInfo} color="blue.500" boxSize={4} /></span>
                            </Tooltip>
                          </HStack>
                          <CurrencyInput
                            value={formData.salvageValue || 0}
                            onChange={(value) => handleInputChange('salvageValue', value)}
                            placeholder="Contoh: Rp 5.000.000"
                            isInvalid={hasFieldError(validationErrors, 'salvageValue')}
                            errorMessage={getFieldError(validationErrors, 'salvageValue')}
                            size="md"
                            min={0}
                            showLabel={false}
                          />
                        </FormControl>
                        
                        <FormControl isInvalid={hasFieldError(validationErrors, 'usefulLife')}>
                          <FormLabel fontSize="sm" fontWeight="medium">Useful Life (Years)</FormLabel>
                          <NumberInput
                            value={formData.usefulLife || 1}
                            onChange={(valueString) => handleInputChange('usefulLife', parseInt(valueString) || 1)}
                            min={1}
                            max={100}
                            size="md"
                          >
                            <NumberInputField />
                            <NumberInputStepper>
                              <NumberIncrementStepper />
                              <NumberDecrementStepper />
                            </NumberInputStepper>
                          </NumberInput>
                          {getFieldError(validationErrors, 'usefulLife') && (
                            <Text fontSize="xs" color="red.500" mt={1}>
                              {getFieldError(validationErrors, 'usefulLife')}
                            </Text>
                          )}
                        </FormControl>
                      </HStack>
                      
                      <HStack w="full" spacing={4}>
                        <FormControl>
                          <HStack spacing={2} align="center">
                            <FormLabel fontSize="sm" fontWeight="medium" mb={0}>Depreciation Method</FormLabel>
                            <Tooltip
                              label={
                                <Box>
                                  <Text fontWeight="semibold" mb={1}>💡 Metode Depresiasi</Text>
                                  <Text fontSize="xs">Straight Line: biaya depresiasi sama tiap periode.</Text>
                                  <Text fontSize="xs">Declining Balance: biaya lebih besar di awal, makin kecil berikutnya.</Text>
                                </Box>
                              }
                              hasArrow
                              placement="top"
                              bg="gray.800"
                              color="white"
                              borderRadius="md"
                              p={3}
                            >
                              <span><Icon as={FiInfo} color="blue.500" boxSize={4} /></span>
                            </Tooltip>
                          </HStack>
                          <Select
                            value={formData.depreciationMethod || 'STRAIGHT_LINE'}
                            onChange={(e) => handleInputChange('depreciationMethod', e.target.value as 'STRAIGHT_LINE' | 'DECLINING_BALANCE')}
                            size="md"
                          >
                            <option value="STRAIGHT_LINE">{DEPRECIATION_METHOD_LABELS.STRAIGHT_LINE}</option>
                            <option value="DECLINING_BALANCE">{DEPRECIATION_METHOD_LABELS.DECLINING_BALANCE}</option>
                          </Select>
                        </FormControl>
                        
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">Asset Status</FormLabel>
                          <Select
                            value={formData.status || 'ACTIVE'}
                            onChange={(e) => handleInputChange('status', e.target.value as 'ACTIVE' | 'INACTIVE' | 'SOLD')}
                            size="md"
                          >
                            <option value="ACTIVE">🟢 Active</option>
                            <option value="INACTIVE">⚪ Inactive</option>
                            <option value="SOLD">🔴 Sold</option>
                          </Select>
                        </FormControl>
                      </HStack>
                      
                      {/* Asset and Depreciation Account Selection */}
                      <HStack w="full" spacing={4}>
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">🏢 Fixed Asset Account</FormLabel>
                          <Select
                            value={formData.assetAccountId || ''}
                            onChange={(e) => handleInputChange('assetAccountId', e.target.value ? parseInt(e.target.value) : undefined)}
                            placeholder="Choose fixed asset account"
                            size="md"
                            isDisabled={isLoadingAccounts}
                          >
                            {fixedAssetAccounts.map((account) => (
                              <option key={account.id} value={account.id}>
                                {account.code} - {account.name} ({formatCurrency(account.balance)})
                              </option>
                            ))}
                          </Select>
                          <Text fontSize="xs" color="gray.500" mt={1}>
                            💡 This account will be debited (default: 1500 - Fixed Assets)
                          </Text>
                        </FormControl>
                        
                        <FormControl>
                          <FormLabel fontSize="sm" fontWeight="medium">📉 Depreciation Expense Account</FormLabel>
                          <Select
                            value={formData.depreciationAccountId || ''}
                            onChange={(e) => handleInputChange('depreciationAccountId', e.target.value ? parseInt(e.target.value) : undefined)}
                            placeholder="Choose depreciation expense account"
                            size="md"
                            isDisabled={isLoadingAccounts}
                          >
                            {depreciationAccounts.map((account) => (
                              <option key={account.id} value={account.id}>
                                {account.code} - {account.name} ({formatCurrency(account.balance)})
                              </option>
                            ))}
                          </Select>
                          <Text fontSize="xs" color="gray.500" mt={1}>
                            💡 For monthly depreciation entries (default: 6201 - Depreciation Expense)
                          </Text>
                        </FormControl>
                      </HStack>
                    </VStack>
                  </Box>

                  {/* Location Information Section */}
                  <Box w="full">
                    <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                      📍 Location Information
                    </Text>
                    <VStack spacing={4}>
                      <FormControl>
                        <FormLabel fontSize="sm" fontWeight="medium">Physical Location</FormLabel>
                        <Input
                          value={formData.location || ''}
                          onChange={(e) => handleInputChange('location', e.target.value)}
                          placeholder="Enter asset physical location (e.g., Office Building Floor 2, Room 201)"
                          size="md"
                        />
                        <Text fontSize="xs" color="gray.500" mt={1}>
                          💡 Describe where this asset is physically located
                        </Text>
                      </FormControl>
                      
                      <FormControl>
                        <FormLabel fontSize="sm" fontWeight="medium">GPS Coordinates (Optional)</FormLabel>
                        <HStack spacing={3}>
                          <Input
                            value={formData.coordinates || ''}
                            onChange={(e) => handleInputChange('coordinates', e.target.value)}
                            placeholder="Click 'Select Location' to choose on map"
                            readOnly
                            flex={1}
                            bg="gray.50"
                            size="md"
                          />
                          <Button
                            leftIcon={<FiMapPin />}
                            onClick={() => setIsMapPickerOpen(true)}
                            colorScheme="blue"
                            variant="outline"
                            size="md"
                            flexShrink={0}
                          >
                            Select on Map
                          </Button>
                          {formData.coordinates && (
                            <Button
                              leftIcon={<FiExternalLink />}
                              onClick={() => assetService.openInMaps(formData.coordinates!)}
                              colorScheme="green"
                              variant="outline"
                              size="md"
                              flexShrink={0}
                            >
                              View
                            </Button>
                          )}
                        </HStack>
                        <Text fontSize="xs" color="gray.500" mt={1}>
                          🗺️ Pinpoint exact location on map for better asset tracking
                        </Text>
                      </FormControl>
                    </VStack>
                  </Box>

                  {/* Asset Image Section */}
                  <Box w="full">
                    <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                      📸 Asset Image
                    </Text>
                    {selectedAsset && selectedAsset.id ? (
                      /* Edit Mode - Full upload functionality */
                      <AssetImageUpload
                        asset={selectedAsset}
                        onImageUpload={handleImageUpload}
                        size="lg"
                        showLabel={false}
                      />
                    ) : (
                      /* Create Mode - Inform user to save first */
                      <Box
                        p={6}
                        border="2px dashed"
                        borderColor="gray.300"
                        borderRadius="lg"
                        textAlign="center"
                        bg="gray.50"
                      >
                        <VStack spacing={3}>
                          <Box
                            p={3}
                            bg="blue.50"
                            borderRadius="full"
                            border="1px"
                            borderColor="blue.100"
                          >
                            <Icon as={FiEdit} boxSize={6} color="blue.500" />
                          </Box>
                          <Text fontSize="md" fontWeight="medium" color="gray.700">
                            Save Asset First to Upload Image
                          </Text>
                          <Text fontSize="sm" color="gray.500" textAlign="center">
                            You can upload an image after creating this asset.
                            Click "Create Asset" button to save, then edit the asset to add an image.
                          </Text>
                        </VStack>
                      </Box>
                    )}
                  </Box>

                  {/* Additional Notes Section */}
                  <Box w="full">
                    <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                      📝 Additional Information
                    </Text>
                    <FormControl>
                      <FormLabel fontSize="sm" fontWeight="medium">Notes</FormLabel>
                      <Textarea
                        value={formData.notes || ''}
                        onChange={(e) => handleInputChange('notes', e.target.value)}
                        placeholder="Add any additional notes, maintenance history, or important information about this asset..."
                        rows={4}
                        resize="vertical"
                        size="md"
                      />
                    </FormControl>
                  </Box>
                </VStack>
              </ModalBody>
              
              <ModalFooter pb={6}>
                <HStack justify="flex-end" spacing={4}>
                  <Button
                    leftIcon={<FiX />}
                    onClick={handleCloseModal}
                    variant="outline"
                  >
                    Cancel
                  </Button>
                  <Button
                    leftIcon={selectedAsset?.id ? <FiEdit /> : <FiPlus />}
                    type="submit"
                    colorScheme="blue"
                    isLoading={isSubmitting}
                    loadingText={selectedAsset?.id ? 'Updating...' : 'Creating...'}
                  >
                    {selectedAsset?.id ? 'Update Asset' : 'Create Asset'}
                  </Button>
                </HStack>
              </ModalFooter>
            </form>
          </ModalContent>
        </Modal>
        
        {/* Interactive Map Picker Modal */}
        <InteractiveMapPicker
          isOpen={isMapPickerOpen}
          onClose={() => setIsMapPickerOpen(false)}
          onLocationSelect={handleLocationPick}
          currentCoordinates={formData.coordinates}
          currentLocationData={{
            name: formData.location ? formData.location.split(' - ')[0] : '',
            description: '',
            address: '',
            coordinates: formData.coordinates || ''
          }}
          title={selectedAsset ? `Select Location - ${selectedAsset.name}` : "Select Asset Location"}
        />

        {/* Image Update Confirmation Dialog */}
        <AlertDialog
          isOpen={isAlertOpen}
          leastDestructiveRef={React.useRef(null)}
          onClose={onAlertClose}
        >
          <AlertDialogOverlay>
            <AlertDialogContent>
              <AlertDialogHeader fontSize="lg" fontWeight="bold">
                Update Asset Image
              </AlertDialogHeader>

              <AlertDialogBody>
                This asset already has an image. Are you sure you want to replace it with the new image?
              </AlertDialogBody>

              <AlertDialogFooter>
                <Button onClick={onAlertClose}>
                  Cancel
                </Button>
                <Button colorScheme="blue" onClick={confirmImageUpdate} ml={3}>
                  Update Image
                </Button>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialogOverlay>
        </AlertDialog>

        {/* Asset Details View Modal */}
        <Modal isOpen={isDetailOpen} onClose={onDetailClose} size="6xl">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              <HStack spacing={3}>
                <Icon as={FiEye} color="blue.500" />
                <Text>Asset Details</Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            
            <ModalBody pb={6}>
              {detailAsset && (
                <VStack spacing={6} align="stretch">
                  {/* Asset Image Section */}
                  <Box textAlign="center">
                    {detailAsset.image_path ? (
                      <Image
                        src={getAssetImageUrl(detailAsset.image_path) || ''}
                        alt={detailAsset.name}
                        maxH="300px"
                        maxW="400px"
                        objectFit="cover"
                        borderRadius="lg"
                        boxShadow="lg"
                        mx="auto"
                        fallback={
                          <Box
                            w="300px"
                            h="200px"
                            bg="gray.100"
                            borderRadius="lg"
                            display="flex"
                            alignItems="center"
                            justifyContent="center"
                            mx="auto"
                          >
                            <Text color="gray.500" fontSize="lg">
                              Failed to load image
                            </Text>
                          </Box>
                        }
                      />
                    ) : (
                      <Box
                        w="300px"
                        h="200px"
                        bg="gray.100"
                        borderRadius="lg"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        mx="auto"
                      >
                        <Text color="gray.500" fontSize="lg">
                          No Image Available
                        </Text>
                      </Box>
                    )}
                  </Box>

                  {/* Basic Information */}
                  <Box>
                    <Text fontSize="xl" fontWeight="bold" mb={4} color="gray.700">
                      📋 Basic Information
                    </Text>
                    <Grid templateColumns="repeat(2, 1fr)" gap={6}>
                      <GridItem>
                        <VStack align="start" spacing={4}>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Asset Code</Text>
                            <Text fontSize="lg" fontWeight="semibold" color="blue.600">
                              {detailAsset.code}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Asset Name</Text>
                            <Text fontSize="md" fontWeight="medium">
                              {detailAsset.name}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Category</Text>
                            <Badge colorScheme="purple" size="lg" px={3} py={1} fontSize="sm">
                              {detailAsset.category}
                            </Badge>
                          </Box>
                        </VStack>
                      </GridItem>
                      <GridItem>
                        <VStack align="start" spacing={4}>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Serial Number</Text>
                            <Text fontSize="md">
                              {detailAsset.serial_number || 'Not specified'}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Condition</Text>
                            <Text fontSize="md">
                              {detailAsset.condition || 'Good'}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Status</Text>
                            <HStack spacing={3}>
                              <Badge colorScheme={getStatusColor(detailAsset.status)} size="lg" px={3} py={1}>
                                {detailAsset.status}
                              </Badge>
                              <Badge colorScheme={detailAsset.is_active ? 'green' : 'red'} size="lg" px={3} py={1}>
                                {detailAsset.is_active ? 'Active' : 'Inactive'}
                              </Badge>
                            </HStack>
                          </Box>
                        </VStack>
                      </GridItem>
                    </Grid>
                  </Box>

              {/* Financial Information */}
              <Box>
                <Text fontSize="xl" fontWeight="bold" mb={4} color="gray.700">
                  💰 Financial Information
                </Text>
                
                {/* Recalculate as of date */}
                <Box mb={4}>
                  <HStack spacing={3} align="center">
                    <Text fontSize="sm" color="gray.600">Recalculate as of</Text>
                    <Input type="date" size="sm" value={recalcDate} onChange={(e) => setRecalcDate(e.target.value)} maxW="200px" />
                    <Button size="sm" colorScheme="blue" isLoading={isRecalcLoading} onClick={handleRecalculate}>Recalculate</Button>
                  </HStack>
                  {recalcResult && (
                    <HStack spacing={6} mt={3}>
                      <Badge colorScheme="teal">Book Value: {formatCurrency(recalcResult.current_book_value)}</Badge>
                      <Badge colorScheme="orange">Accumulated: {formatCurrency(recalcResult.accumulated_depreciation)}</Badge>
                      <Badge colorScheme="purple">Method: {recalcResult.depreciation_method.replace('_',' ')}</Badge>
                    </HStack>
                  )}
                </Box>
                    <Grid templateColumns="repeat(2, 1fr)" gap={6}>
                      <GridItem>
                        <VStack align="start" spacing={4}>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Purchase Date</Text>
                            <Text fontSize="md">
                              {new Date(detailAsset.purchase_date).toLocaleDateString('id-ID', {
                                year: 'numeric',
                                month: 'long', 
                                day: 'numeric'
                              })}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Purchase Price</Text>
                            <Text fontSize="lg" fontWeight="semibold" color="green.600">
                              {formatCurrency(detailAsset.purchase_price)}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Salvage Value</Text>
                            <Text fontSize="md">
                              {formatCurrency(detailAsset.salvage_value)}
                            </Text>
                          </Box>
                        </VStack>
                      </GridItem>
                      <GridItem>
                        <VStack align="start" spacing={4}>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Useful Life</Text>
                            <Text fontSize="md">
                              {detailAsset.useful_life} years
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Accumulated Depreciation</Text>
                            <Text fontSize="lg" fontWeight="semibold" color="orange.600">
                              {formatCurrency(detailAsset.accumulated_depreciation)}
                            </Text>
                          </Box>
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Current Book Value</Text>
                            <Text fontSize="lg" fontWeight="bold" color="blue.600">
                              {formatCurrency(calculateBookValue(detailAsset))}
                            </Text>
                          </Box>
                        </VStack>
                      </GridItem>
                    </Grid>
                    <Box mt={4}>
                      <Text fontSize="sm" color="gray.500" fontWeight="medium">Depreciation Method</Text>
                      <Badge colorScheme="teal" size="lg" px={3} py={1}>
                        {DEPRECIATION_METHOD_LABELS[detailAsset.depreciation_method as keyof typeof DEPRECIATION_METHOD_LABELS]}
                      </Badge>
                    </Box>
                  </Box>

                  {/* Location Information */}
                  {(detailAsset.location || detailAsset.coordinates) && (
                    <Box>
                      <Text fontSize="xl" fontWeight="bold" mb={4} color="gray.700">
                        📍 Location Information
                      </Text>
                      <VStack align="start" spacing={3}>
                        {detailAsset.location && (
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">Physical Location</Text>
                            <Text fontSize="md">
                              {detailAsset.location}
                            </Text>
                          </Box>
                        )}
                        {detailAsset.coordinates && (
                          <Box>
                            <Text fontSize="sm" color="gray.500" fontWeight="medium">GPS Coordinates</Text>
                            <HStack spacing={3}>
                              <Text fontSize="md" fontFamily="mono">
                                {detailAsset.coordinates}
                              </Text>
                              <Button
                                size="sm"
                                leftIcon={<FiExternalLink />}
                                onClick={() => assetService.openInMaps(detailAsset.coordinates!)}
                                colorScheme="blue"
                                variant="outline"
                              >
                                View on Map
                              </Button>
                            </HStack>
                          </Box>
                        )}
                      </VStack>
                    </Box>
                  )}

                  {/* Additional Notes */}
                  {detailAsset.notes && (
                    <Box>
                      <Text fontSize="xl" fontWeight="bold" mb={4} color="gray.700">
                        📝 Notes
                      </Text>
                      <Box
                        p={4}
                        bg="gray.50"
                        borderRadius="lg"
                        border="1px solid"
                        borderColor="gray.200"
                      >
                        <Text fontSize="md" whiteSpace="pre-wrap">
                          {detailAsset.notes}
                        </Text>
                      </Box>
                    </Box>
                  )}

                  {/* Timestamps */}
                  <Box>
                    <Text fontSize="lg" fontWeight="semibold" mb={3} color="gray.600">
                      📅 Record Information
                    </Text>
                    <Grid templateColumns="repeat(2, 1fr)" gap={4}>
                      <GridItem>
                        <Box>
                          <Text fontSize="sm" color="gray.500" fontWeight="medium">Created At</Text>
                          <Text fontSize="sm">
                            {new Date(detailAsset.created_at).toLocaleString('id-ID')}
                          </Text>
                        </Box>
                      </GridItem>
                      <GridItem>
                        <Box>
                          <Text fontSize="sm" color="gray.500" fontWeight="medium">Last Updated</Text>
                          <Text fontSize="sm">
                            {new Date(detailAsset.updated_at).toLocaleString('id-ID')}
                          </Text>
                        </Box>
                      </GridItem>
                    </Grid>
                  </Box>
                </VStack>
              )}
            </ModalBody>
            
            <ModalFooter>
              <HStack spacing={3}>
                <Button
                  leftIcon={<FiEdit />}
                  onClick={() => {
                    if (detailAsset) {
                      onDetailClose();
                      handleEdit(detailAsset);
                    }
                  }}
                  colorScheme="blue"
                  variant="outline"
                >
                  Edit Asset
                </Button>
                <Button onClick={onDetailClose}>
                  Close
                </Button>
              </HStack>
            </ModalFooter>
          </ModalContent>
        </Modal>

        {/* Category Management Modal */}
        <Modal isOpen={isCategoryModalOpen} onClose={handleCloseCategoryModal} size="lg">
          <ModalOverlay />
          <ModalContent>
            <ModalHeader>
              <HStack spacing={3}>
                <Icon as={FiSettings} color="gray.500" />
                <Text>Manage Asset Categories</Text>
              </HStack>
            </ModalHeader>
            <ModalCloseButton />
            
            <ModalBody pb={6}>
              <VStack spacing={6} align="stretch">
                {/* Add New Category Section */}
                <Box>
                  <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                    ➕ Add New Category
                  </Text>
                  <HStack spacing={3}>
                    <FormControl flex={1}>
                      <Input
                        value={newCategoryName}
                        onChange={(e) => setNewCategoryName(e.target.value)}
                        placeholder="Enter category name"
                      />
                    </FormControl>
                    <FormControl w="160px">
                      <Input
                        value={newCategoryCode}
                        onChange={(e) => setNewCategoryCode(e.target.value.toUpperCase())}
                        placeholder="Code"
                        maxLength={10}
                      />
                    </FormControl>
                    {editingCategoryIndex !== null ? (
                      <>
                        <Button
                          colorScheme="blue"
                          onClick={handleUpdateCategory}
                          isDisabled={!newCategoryName.trim()}
                        >
                          Update
                        </Button>
                        <Button
                          variant="outline"
                          onClick={cancelEdit}
                        >
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <Button
                        colorScheme="blue"
                        leftIcon={<FiPlus />}
                        onClick={handleAddCategory}
                        isDisabled={!newCategoryName.trim()}
                      >
                        Add
                      </Button>
                    )}
                  </HStack>
                  <Text fontSize="xs" color="gray.500" mt={2}>
                    💡 Provide a short code/prefix (e.g., RE, IT, VEH). This prefix will be used for asset codes.
                  </Text>
                </Box>

                {/* Existing Categories List */}
                <Box>
                  <Text fontSize="md" fontWeight="semibold" color="gray.700" mb={4}>
                    📋 Existing Categories ({customCategories.length})
                  </Text>
                  <VStack spacing={2} align="stretch" maxH="300px" overflowY="auto">
                    {customCategories.map((category, index) => {
                      const isUsed = assets.some(asset => asset.category === category);
                      const isDefault = ASSET_CATEGORIES.includes(category as any);
                      
                      return (
                        <HStack
                          key={index}
                          p={3}
                          bg={editingCategoryIndex === index ? 'blue.50' : isDefault ? 'gray.50' : 'white'}
                          border="1px solid"
                          borderColor={editingCategoryIndex === index ? 'blue.200' : 'gray.200'}
                          borderRadius="md"
                          justify="space-between"
                        >
                          <HStack spacing={3} flex={1}>
                            <Text
                              fontSize="sm"
                              fontWeight={editingCategoryIndex === index ? 'semibold' : 'normal'}
                              color={editingCategoryIndex === index ? 'blue.700' : 'gray.700'}
                            >
                              {category}
                            </Text>
                            <Badge colorScheme="purple" size="sm" fontSize="xs">
                              {getCodeForCategory(category) || '—'}
                            </Badge>
                            <HStack spacing={2}>
                              {isDefault && (
                                <Badge colorScheme="gray" size="sm" fontSize="xs">
                                  Default
                                </Badge>
                              )}
                              {isUsed && (
                                <Badge colorScheme="green" size="sm" fontSize="xs">
                                  In Use
                                </Badge>
                              )}
                            </HStack>
                          </HStack>
                          
                          <HStack spacing={2}>
                            <Button
                              size="sm"
                              variant="ghost"
                              leftIcon={<FiEdit />}
                              onClick={() => handleEditCategory(index)}
                              isDisabled={editingCategoryIndex !== null && editingCategoryIndex !== index}
                            >
                              Edit
                            </Button>
                            <Button
                              size="sm"
                              variant="ghost"
                              colorScheme="red"
                              leftIcon={<FiTrash2 />}
                              onClick={() => handleDeleteCategory(index)}
                              isDisabled={isUsed || isDefault}
                              title={isUsed ? 'Cannot delete: Category is in use' : isDefault ? 'Cannot delete: Default category' : ''}
                            >
                              Delete
                            </Button>
                          </HStack>
                        </HStack>
                      );
                    })}
                    
                    {customCategories.length === 0 && (
                      <Box
                        p={6}
                        textAlign="center"
                        bg="gray.50"
                        borderRadius="md"
                        border="2px dashed"
                        borderColor="gray.300"
                      >
                        <Text color="gray.500">
                          No categories available. Add your first category above.
                        </Text>
                      </Box>
                    )}
                  </VStack>
                </Box>
              </VStack>
            </ModalBody>
            
            <ModalFooter>
              <Button onClick={handleCloseCategoryModal}>
                Close
              </Button>
            </ModalFooter>
          </ModalContent>
        </Modal>
      </Box>
    </SimpleLayout>
  );
};

export default AssetsPage;
