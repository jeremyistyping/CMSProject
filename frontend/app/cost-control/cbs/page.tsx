'use client';

import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Heading,
  Button,
  HStack,
  VStack,
  useToast,
  Select,
  Text,
  Spinner,
  Container,
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
} from '@chakra-ui/react';
import { FiPlus, FiRefreshCw } from 'react-icons/fi';
import SimpleLayout from '../../../src/components/layout/SimpleLayout';
import CBSTreeView from '../../../src/components/cost-control/CBSTreeView';
import CBSNodeModal from '../../../src/components/cost-control/CBSNodeModal';
import cbsService from '../../../src/services/cbsService';
import projectService from '../../../src/services/projectService';
import { CBSNode } from '../../../src/types/cbs';
import { Project } from '../../../src/types/project';
import { useModulePermissions } from '../../../src/hooks/usePermissions';

export default function CBSPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<number | null>(null);
  const [cbsNodes, setCbsNodes] = useState<CBSNode[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  // Modal State
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedParentId, setSelectedParentId] = useState<number | undefined>(undefined);
  const [nodeToEdit, setNodeToEdit] = useState<CBSNode | undefined>(undefined);

  const toast = useToast();
  const { canCreate, canEdit, canDelete } = useModulePermissions('cbs');

  // Fetch Projects
  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const data = await projectService.getAllProjects();
        setProjects(data);
        if (data.length > 0) {
          setSelectedProjectId(data[0].id);
        }
      } catch (error) {
        console.error('Error fetching projects:', error);
        toast({
          title: 'Error',
          description: 'Failed to load projects',
          status: 'error',
          duration: 3000,
        });
      }
    };
    fetchProjects();
  }, [toast]);

  // Fetch CBS Tree
  const fetchCBSTree = useCallback(async () => {
    if (!selectedProjectId) return;

    setIsLoading(true);
    try {
      const data = await cbsService.getCBSTree(selectedProjectId);
      setCbsNodes(data);
    } catch (error) {
      console.error('Error fetching CBS tree:', error);
      toast({
        title: 'Error',
        description: 'Failed to load CBS tree',
        status: 'error',
        duration: 3000,
      });
    } finally {
      setIsLoading(false);
    }
  }, [selectedProjectId, toast]);

  useEffect(() => {
    fetchCBSTree();
  }, [fetchCBSTree]);

  // Handlers
  const handleAddNode = (parentId?: number) => {
    console.log('handleAddNode called!', { parentId, selectedProjectId, isModalOpen });
    setSelectedParentId(parentId);
    setNodeToEdit(undefined);
    setIsModalOpen(true);
    console.log('handleAddNode: set isModalOpen to true');
  };

  const handleEditNode = (node: CBSNode) => {
    setSelectedParentId(node.parent_id);
    setNodeToEdit(node);
    setIsModalOpen(true);
  };

  const handleDeleteNode = async (node: CBSNode) => {
    if (!window.confirm(`Are you sure you want to delete ${node.name}? This will also delete all sub-nodes.`)) {
      return;
    }

    try {
      await cbsService.deleteCBSNode(node.id);
      toast({
        title: 'Success',
        description: 'CBS Node deleted successfully',
        status: 'success',
        duration: 3000,
      });
      fetchCBSTree();
    } catch (error: any) {
      toast({
        title: 'Error',
        description: error.response?.data?.error || 'Failed to delete node',
        status: 'error',
        duration: 3000,
      });
    }
  };

  return (
    <SimpleLayout allowedRoles={['cost_control', 'admin', 'director', 'gm', 'purchasing']}>
      <Container maxW="container.xl" py={6}>
        <VStack spacing={6} align="stretch">
          {/* Header & Controls */}
          <HStack justify="space-between">
            <Breadcrumb>
              <BreadcrumbItem>
                <BreadcrumbLink href="/cost-control">Cost Control</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbItem isCurrentPage>
                <BreadcrumbLink>CBS</BreadcrumbLink>
              </BreadcrumbItem>
            </Breadcrumb>

            <HStack>
              <Select
                w="300px"
                value={selectedProjectId || ''}
                onChange={(e) => setSelectedProjectId(Number(e.target.value))}
              >
                {projects.map(p => (
                  <option key={p.id} value={p.id}>{p.project_name}</option>
                ))}
              </Select>
              <Button
                leftIcon={<FiRefreshCw />}
                onClick={fetchCBSTree}
                isLoading={isLoading}
              >
                Refresh
              </Button>
              <Button
                leftIcon={<FiPlus />}
                colorScheme="blue"
                onClick={() => {
                  console.log('Create Root Node clicked!', { selectedProjectId, isModalOpen, canCreate });
                  handleAddNode();
                }}
                isDisabled={!selectedProjectId}
              >
                Create Root Node
              </Button>
            </HStack>
          </HStack>

          {/* Main Content */}
          <Box bg="white" p={6} borderRadius="lg" shadow="sm" minH="500px">
            {isLoading ? (
              <VStack justify="center" h="200px">
                <Spinner size="xl" />
                <Text>Loading CBS Structure...</Text>
              </VStack>
            ) : (
              <CBSTreeView
                nodes={cbsNodes}
                onAddNode={canCreate ? handleAddNode : () => { }}
                onEditNode={canEdit ? handleEditNode : () => { }}
                onDeleteNode={canDelete ? handleDeleteNode : () => { }}
              />
            )}
          </Box>
        </VStack>

        {/* Modal */}
        {selectedProjectId !== null && (
          <CBSNodeModal
            isOpen={isModalOpen}
            onClose={() => setIsModalOpen(false)}
            onSuccess={fetchCBSTree}
            projectId={selectedProjectId}
            parentId={selectedParentId}
            nodeToEdit={nodeToEdit}
          />
        )}
      </Container>
    </SimpleLayout>
  );
}
