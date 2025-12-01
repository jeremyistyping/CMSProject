'use client';

import React, { useState, useEffect } from 'react';
import SimpleLayout from '@/components/layout/SimpleLayout';
import { useModulePermissions } from '@/hooks/usePermissions';
import {
  Box,
  Heading,
  Text,
  VStack,
  HStack,
  Spinner,
  Alert,
  AlertIcon,
  useColorModeValue,
  Button,
  useDisclosure,
  Select,
  Input,
  InputGroup,
  InputLeftElement,
  useToast,
  AlertDialog,
  AlertDialogBody,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogContent,
  AlertDialogOverlay,
} from '@chakra-ui/react';
import { FiPlus, FiSearch, FiFilter } from 'react-icons/fi';
import PRList from '@/components/cost-control/PRList';
import CreatePRModal from '@/components/cost-control/CreatePRModal';
import PRDetailModal from '@/components/cost-control/PRDetailModal';
import purchaseRequestService from '@/services/purchaseRequestService';
import projectService from '@/services/projectService';
import { PurchaseRequest } from '@/types/purchaseRequest';
import { Project } from '@/types/project';

const PurchaseRequestsPage: React.FC = () => {
  const { canView, canCreate, canEdit, canDelete, loading: permLoading } = useModulePermissions('purchases');
  const headingColor = useColorModeValue('gray.800', 'gray.100');
  const textColor = useColorModeValue('gray.600', 'gray.300');
  const boxBg = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.700');

  // State
  const [purchaseRequests, setPurchaseRequests] = useState<PurchaseRequest[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedProject, setSelectedProject] = useState<string>('');
  const [selectedStatus, setSelectedStatus] = useState<string>('');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedPR, setSelectedPR] = useState<PurchaseRequest | null>(null);
  const [prToEdit, setPrToEdit] = useState<PurchaseRequest | null>(null);
  const [prToDelete, setPrToDelete] = useState<PurchaseRequest | null>(null);
  const [isDeleteAlertOpen, setIsDeleteAlertOpen] = useState(false);
  const cancelRef = React.useRef(null);
  const toast = useToast();

  // Modals
  const {
    isOpen: isCreateOpen,
    onOpen: onCreateOpen,
    onClose: onCreateClose
  } = useDisclosure();

  const {
    isOpen: isDetailOpen,
    onOpen: onDetailOpen,
    onClose: onDetailClose
  } = useDisclosure();

  useEffect(() => {
    if (canView) {
      fetchData();
    }
  }, [canView]);

  useEffect(() => {
    if (canView) {
      fetchPRs();
    }
  }, [selectedProject, selectedStatus]);

  const fetchData = async () => {
    try {
      const [projectsData] = await Promise.all([
        projectService.getAllProjects(),
      ]);
      setProjects(projectsData);
      fetchPRs();
    } catch (error) {
      console.error('Error fetching initial data:', error);
    }
  };

  const fetchPRs = async () => {
    try {
      setIsLoading(true);
      const filter: any = {};
      if (selectedProject) filter.project_id = Number(selectedProject);
      if (selectedStatus) filter.status = selectedStatus;

      const data = await purchaseRequestService.getAll(filter);
      setPurchaseRequests(data);
    } catch (error) {
      console.error('Error fetching PRs:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleViewPR = (pr: PurchaseRequest) => {
    setSelectedPR(pr);
    onDetailOpen();
  };

  const handleEditPR = (pr: PurchaseRequest) => {
    setPrToEdit(pr);
    onCreateOpen();
  };

  const handleDeletePR = (pr: PurchaseRequest) => {
    setPrToDelete(pr);
    setIsDeleteAlertOpen(true);
  };

  const confirmDeletePR = async () => {
    if (!prToDelete) return;

    try {
      await purchaseRequestService.delete(prToDelete.id);
      toast({
        title: 'Success',
        description: 'Purchase Request deleted successfully',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      fetchPRs();
    } catch (error) {
      console.error('Error deleting PR:', error);
      toast({
        title: 'Error',
        description: 'Failed to delete Purchase Request',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsDeleteAlertOpen(false);
      setPrToDelete(null);
    }
  };

  const handleCreateClose = () => {
    setPrToEdit(null);
    onCreateClose();
  };

  const filteredPRs = purchaseRequests.filter(pr =>
    pr.code.toLowerCase().includes(searchQuery.toLowerCase()) ||
    pr.project?.project_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    pr.requester?.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  if (permLoading) {
    return (
      <SimpleLayout>
        <Box display="flex" alignItems="center" justifyContent="center" minH="60vh">
          <HStack spacing={3}>
            <Spinner />
            <Text>Checking permissions...</Text>
          </HStack>
        </Box>
      </SimpleLayout>
    );
  }

  if (!canView) {
    return (
      <SimpleLayout>
        <Box maxW="xl">
          <Alert status="error" borderRadius="md">
            <AlertIcon />
            <Box>
              <Heading size="sm" mb={1}>Access Denied</Heading>
              <Text fontSize="sm">Anda tidak memiliki akses ke modul Cost Control. Silakan hubungi administrator.</Text>
            </Box>
          </Alert>
        </Box>
      </SimpleLayout>
    );
  }

  return (
    <SimpleLayout>
      <Box>
        <VStack align="start" spacing={6} mb={8}>
          <HStack justify="space-between" w="full">
            <VStack align="start" spacing={1}>
              <Heading size="lg" color={headingColor}>Purchase Request Management</Heading>
              <Text fontSize="sm" color={textColor}>
                Manage and approve purchase requests for your projects
              </Text>
            </VStack>
            {canCreate && (
              <Button leftIcon={<FiPlus />} colorScheme="blue" onClick={() => { setPrToEdit(null); onCreateOpen(); }}>
                Create New PR
              </Button>
            )}
          </HStack>

          {/* Filters */}
          <HStack w="full" spacing={4} bg={boxBg} p={4} borderRadius="lg" borderWidth="1px" borderColor={borderColor}>
            <InputGroup maxW="300px">
              <InputLeftElement pointerEvents="none">
                <FiSearch color="gray.300" />
              </InputLeftElement>
              <Input
                placeholder="Search by code, project, or requester..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </InputGroup>

            <Select
              placeholder="All Projects"
              maxW="200px"
              value={selectedProject}
              onChange={(e) => setSelectedProject(e.target.value)}
            >
              {projects.map(p => (
                <option key={p.id} value={p.id}>{p.project_name}</option>
              ))}
            </Select>

            <Select
              placeholder="All Status"
              maxW="200px"
              value={selectedStatus}
              onChange={(e) => setSelectedStatus(e.target.value)}
            >
              <option value="PENDING">Pending</option>
              <option value="APPROVED">Approved</option>
              <option value="REJECTED">Rejected</option>
              <option value="REVISION">Revision</option>
              <option value="PO_CREATED">PO Created</option>
            </Select>
          </HStack>
        </VStack>

        <Box
          bg={boxBg}
          borderWidth="1px"
          borderColor={borderColor}
          borderRadius="lg"
          p={6}
          minH="400px"
        >
          {isLoading ? (
            <Box display="flex" justifyContent="center" py={10}>
              <Spinner size="xl" color="blue.500" />
            </Box>
          ) : (
            <PRList
              purchaseRequests={filteredPRs}
              onView={handleViewPR}
              onEdit={canEdit ? handleEditPR : undefined}
              onDelete={canDelete ? handleDeletePR : undefined}
            />
          )}
        </Box>
      </Box>

      {/* Modals */}
      <CreatePRModal
        isOpen={isCreateOpen}
        onClose={handleCreateClose}
        onSuccess={fetchPRs}
        prToEdit={prToEdit}
      />

      <PRDetailModal
        isOpen={isDetailOpen}
        onClose={onDetailClose}
        pr={selectedPR}
        onUpdate={fetchPRs}
      />

      <AlertDialog
        isOpen={isDeleteAlertOpen}
        leastDestructiveRef={cancelRef}
        onClose={() => setIsDeleteAlertOpen(false)}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Delete Purchase Request
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure you want to delete this Purchase Request? This action cannot be undone.
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={() => setIsDeleteAlertOpen(false)}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={confirmDeletePR} ml={3}>
                Delete
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </SimpleLayout>
  );
};

export default PurchaseRequestsPage;

