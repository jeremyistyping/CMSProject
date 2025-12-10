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
import PRVerificationModal from '@/components/cost-control/PRVerificationModal';
import CreatePOModal from '@/components/cost-control/CreatePOModal';
import GoodsReceiptModal from '@/components/cost-control/GoodsReceiptModal';
import purchaseRequestService from '@/services/purchaseRequestService';
import purchaseOrderService from '@/services/purchaseOrderService';
import projectService from '@/services/projectService';
import cbsService from '@/services/cbsService';
import { PurchaseRequest } from '@/types/purchaseRequest';
import { Project } from '@/types/project';
import { CBSNode } from '@/types/cbs';

const PurchaseRequestsPage: React.FC = () => {
  // Approve/Reject State
  const [isApproveAlertOpen, setIsApproveAlertOpen] = useState(false);
  const [isRejectAlertOpen, setIsRejectAlertOpen] = useState(false);
  const [prToApprove, setPrToApprove] = useState<PurchaseRequest | null>(null);
  const [prToReject, setPrToReject] = useState<PurchaseRequest | null>(null);
  const [rejectionReason, setRejectionReason] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);

  // Verification State
  const [prToVerify, setPrToVerify] = useState<PurchaseRequest | null>(null);
  const [cbsNodes, setCbsNodes] = useState<CBSNode[]>([]);

  // PO State
  const [prForPO, setPrForPO] = useState<PurchaseRequest | null>(null);
  const [prForGR, setPrForGR] = useState<PurchaseRequest | null>(null);

  const { canView, canCreate, canEdit, canDelete, canApprove, loading: permLoading } = useModulePermissions('purchases');
  // Check if user has cost control permissions for verification
  const { canCreate: canVerify } = useModulePermissions('cbs');

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

  const {
    isOpen: isVerifyOpen,
    onOpen: onVerifyOpen,
    onClose: onVerifyClose
  } = useDisclosure();

  const {
    isOpen: isPOOpen,
    onOpen: onPOOpen,
    onClose: onPOClose
  } = useDisclosure();

  const {
    isOpen: isGROpen,
    onOpen: onGROpen,
    onClose: onGRClose
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

  const handleApprovePR = (pr: PurchaseRequest) => {
    setPrToApprove(pr);
    setIsApproveAlertOpen(true);
  };

  const handleRejectPR = (pr: PurchaseRequest) => {
    setPrToReject(pr);
    setRejectionReason('');
    setIsRejectAlertOpen(true);
  };

  const handleVerifyPR = async (pr: PurchaseRequest) => {
    if (!pr.project_id) {
      toast({
        title: 'Error',
        description: 'PR does not have a project assigned',
        status: 'error',
        duration: 3000,
      });
      return;
    }

    try {
      // Fetch CBS nodes for the project
      const nodes = await cbsService.getCBSTree(pr.project_id);
      setCbsNodes(nodes);
      setPrToVerify(pr);
      onVerifyOpen();
    } catch (error) {
      console.error('Error fetching CBS nodes:', error);
      toast({
        title: 'Error',
        description: 'Failed to load CBS structure for verification',
        status: 'error',
        duration: 3000,
      });
    }
  };

  const handleCreatePO = (pr: PurchaseRequest) => {
    setPrForPO(pr);
    onPOOpen();
  };

  const handleReceiveGoods = (pr: PurchaseRequest) => {
    setPrForGR(pr);
    onGROpen();
  };

  const handleDownloadPO = async (pr: PurchaseRequest) => {
    try {
      // Get PO ID from PR
      const pos = await purchaseOrderService.getByPRId(pr.id);
      if (pos && pos.length > 0) {
        const activePO = pos.find(po => po.status !== 'CANCELLED');
        if (activePO) {
          await purchaseOrderService.downloadPOPDF(activePO.id);
          toast({
            title: 'Success',
            description: 'PO PDF downloaded successfully',
            status: 'success',
            duration: 2000,
          });
        }
      }
    } catch (error) {
      console.error('Error downloading PO PDF:', error);
      toast({
        title: 'Error',
        description: 'Failed to download PO PDF',
        status: 'error',
        duration: 3000,
      });
    }
  };

  const handleDownloadGR = async (pr: PurchaseRequest) => {
    try {
      // Get PO ID from PR
      const pos = await purchaseOrderService.getByPRId(pr.id);
      if (pos && pos.length > 0) {
        const activePO = pos.find(po => po.status !== 'CANCELLED');
        if (activePO) {
          await purchaseOrderService.downloadGRPDF(activePO.id);
          toast({
            title: 'Success',
            description: 'Goods Receipt PDF downloaded successfully',
            status: 'success',
            duration: 2000,
          });
        }
      }
    } catch (error) {
      console.error('Error downloading GR PDF:', error);
      toast({
        title: 'Error',
        description: 'Failed to download Goods Receipt PDF',
        status: 'error',
        duration: 3000,
      });
    }
  };

  const confirmApprovePR = async () => {
    if (!prToApprove) return;

    setIsProcessing(true);
    try {
      await purchaseRequestService.updateStatus(prToApprove.id, 'APPROVED');
      toast({
        title: 'Purchase Request Approved',
        description: `PR ${prToApprove.code} has been approved successfully.`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      fetchPRs();
    } catch (error) {
      console.error('Error approving PR:', error);
      toast({
        title: 'Error',
        description: 'Failed to approve Purchase Request',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsProcessing(false);
      setIsApproveAlertOpen(false);
      setPrToApprove(null);
    }
  };

  const confirmRejectPR = async () => {
    if (!prToReject) return;
    if (!rejectionReason.trim()) {
      toast({
        title: 'Error',
        description: 'Please provide a reason for rejection',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
      return;
    }

    setIsProcessing(true);
    try {
      await purchaseRequestService.updateStatus(prToReject.id, 'REJECTED', rejectionReason);
      toast({
        title: 'Purchase Request Rejected',
        description: `PR ${prToReject.code} has been rejected.`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      });
      fetchPRs();
    } catch (error) {
      console.error('Error rejecting PR:', error);
      toast({
        title: 'Error',
        description: 'Failed to reject Purchase Request',
        status: 'error',
        duration: 3000,
        isClosable: true,
      });
    } finally {
      setIsProcessing(false);
      setIsRejectAlertOpen(false);
      setPrToReject(null);
      setRejectionReason('');
    }
  };

  const handleCreateClose = () => {
    setPrToEdit(null);
    onCreateClose();
  };

  const filteredPRs = purchaseRequests.filter(pr =>
    pr.code.toLowerCase().includes(searchQuery.toLowerCase()) ||
    pr.project?.project_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    `${pr.requester?.first_name || ''} ${pr.requester?.last_name || ''}`.toLowerCase().includes(searchQuery.toLowerCase())
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
              <option value="PENDING_VERIFICATION">Pending Verification</option>
              <option value="VERIFIED">Verified</option>
              <option value="APPROVED">Approved</option>
              <option value="REJECTED">Rejected</option>
              <option value="REVISION">Revision</option>
              <option value="PO_CREATED">PO Created</option>
              <option value="COMPLETED">Completed</option>
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
              onApprove={canApprove ? handleApprovePR : undefined}
              onReject={canApprove ? handleRejectPR : undefined}
              onVerify={canVerify ? handleVerifyPR : undefined}
              onCreatePO={canCreate ? handleCreatePO : undefined}
              onReceiveGoods={canCreate ? handleReceiveGoods : undefined}
              onDownloadPO={canView ? handleDownloadPO : undefined}
              onDownloadGR={canView ? handleDownloadGR : undefined}
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

      <PRVerificationModal
        isOpen={isVerifyOpen}
        onClose={onVerifyClose}
        pr={prToVerify}
        cbsNodes={cbsNodes}
        onSuccess={fetchPRs}
      />

      <CreatePOModal
        isOpen={isPOOpen}
        onClose={onPOClose}
        pr={prForPO}
        onSuccess={fetchPRs}
      />

      <GoodsReceiptModal
        isOpen={isGROpen}
        onClose={onGRClose}
        pr={prForGR}
        onSuccess={fetchPRs}
      />

      {/* Delete Alert */}
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

      {/* Approve Alert */}
      <AlertDialog
        isOpen={isApproveAlertOpen}
        leastDestructiveRef={cancelRef}
        onClose={() => setIsApproveAlertOpen(false)}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Approve Purchase Request
            </AlertDialogHeader>

            <AlertDialogBody>
              Are you sure you want to approve Purchase Request <strong>{prToApprove?.code}</strong>?
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={() => setIsApproveAlertOpen(false)}>
                Cancel
              </Button>
              <Button colorScheme="green" onClick={confirmApprovePR} ml={3} isLoading={isProcessing}>
                Approve
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>

      {/* Reject Alert */}
      <AlertDialog
        isOpen={isRejectAlertOpen}
        leastDestructiveRef={cancelRef}
        onClose={() => setIsRejectAlertOpen(false)}
      >
        <AlertDialogOverlay>
          <AlertDialogContent>
            <AlertDialogHeader fontSize="lg" fontWeight="bold">
              Reject Purchase Request
            </AlertDialogHeader>

            <AlertDialogBody>
              <VStack align="start" spacing={3}>
                <Text>Are you sure you want to reject Purchase Request <strong>{prToReject?.code}</strong>?</Text>
                <Text fontSize="sm" fontWeight="bold">Reason for Rejection:</Text>
                <Input
                  placeholder="Enter reason..."
                  value={rejectionReason}
                  onChange={(e) => setRejectionReason(e.target.value)}
                />
              </VStack>
            </AlertDialogBody>

            <AlertDialogFooter>
              <Button ref={cancelRef} onClick={() => setIsRejectAlertOpen(false)}>
                Cancel
              </Button>
              <Button colorScheme="red" onClick={confirmRejectPR} ml={3} isLoading={isProcessing}>
                Reject
              </Button>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialogOverlay>
      </AlertDialog>
    </SimpleLayout>
  );
};

export default PurchaseRequestsPage;
